package workout

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"membership-gym/internal/middleware"
	"membership-gym/pkg/datetime"
	"membership-gym/pkg/day"
	"membership-gym/pkg/duration"
	apperror "membership-gym/pkg/errors"
	"membership-gym/pkg/pagination"
	"membership-gym/pkg/token"
	txtx "membership-gym/pkg/tx"
)

type resolvedQR struct {
	GymID        int64
	MemberID     int64
	MemberPublic uuid.UUID
	MemberName   string
	MemberCode   string
	GymName      string
	Timezone     string
}

func (s *Service) resolveQR(ctx context.Context, qrToken string) (*resolvedQR, error) {
	rows, err := s.repo.V2ResolveQR(ctx, qrToken)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, apperror.NotFound("QR tidak valid")
	}
	if fmt.Sprint(rows[0]["member_status"]) != "active" {
		return nil, apperror.BadRequest("Member tidak aktif")
	}
	return &resolvedQR{
		GymID: rows[0]["gym_id"].(int64), MemberID: rows[0]["member_id"].(int64),
		MemberPublic: rows[0]["member_public_id"].(uuid.UUID), MemberName: fmt.Sprint(rows[0]["member_name"]),
		MemberCode: fmt.Sprint(rows[0]["member_code"]), GymName: fmt.Sprint(rows[0]["gym_name"]), Timezone: fmt.Sprint(rows[0]["timezone"]),
	}, nil
}

func normalizeCode(code string) string {
	code = strings.ToLower(strings.TrimSpace(code))
	code = strings.ReplaceAll(code, " ", "_")
	code = strings.ReplaceAll(code, "-", "_")
	return code
}

func (s *Service) CreateMuscleGroup(ctx context.Context, auth middleware.AuthUser, req CreateMuscleGroupRequest) (map[string]any, error) {
	var rows []map[string]any
	err := txtx.WithTx(ctx, s.repo.DB(), func(tx pgx.Tx) error {
		var e error
		rows, e = s.repo.InsertMuscleGroup(ctx, tx, token.GeneratePublicID(), auth.GymID, req.Name, normalizeCode(req.Code), req.Description)
		if e != nil {
			return e
		}
		return s.repo.LogAudit(ctx, tx, auth.GymID, &auth.UserID, "muscle_group.created", "muscle_groups", nil, rows[0])
	})
	if err != nil {
		return nil, err
	}
	return rows[0], nil
}

func (s *Service) CreateWorkoutTemplate(ctx context.Context, auth middleware.AuthUser, req CreateWorkoutTemplateRequest) (map[string]any, error) {
	var rows []map[string]any
	err := txtx.WithTx(ctx, s.repo.DB(), func(tx pgx.Tx) error {
		var e error
		rows, e = s.repo.InsertWorkoutTemplate(ctx, tx, token.GeneratePublicID(), auth.GymID, req.Name, req.Description, auth.UserID)
		if e != nil {
			return e
		}
		return s.repo.LogAudit(ctx, tx, auth.GymID, &auth.UserID, "workout_template.created", "workout_templates", nil, rows[0])
	})
	if err != nil {
		return nil, err
	}
	return rows[0], nil
}

func (s *Service) TemplateDays(ctx context.Context, auth middleware.AuthUser, templatePublicID string) ([]map[string]any, error) {
	templateID, err := parsePublicID(templatePublicID)
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.TemplateDays(ctx, auth.GymID, templateID)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		row["day_name"] = day.DayOfWeekName(intNum(row["day_of_week"]))
		if row["muscle_group_public_id"] != nil {
			row["muscle_group"] = map[string]any{"public_id": row["muscle_group_public_id"], "name": row["muscle_group_name"], "code": row["muscle_group_code"]}
		}
		delete(row, "muscle_group_public_id")
		delete(row, "muscle_group_name")
		delete(row, "muscle_group_code")
	}
	return rows, nil
}

func (s *Service) ReplaceTemplateDays(ctx context.Context, auth middleware.AuthUser, templatePublicID string, req ReplaceWorkoutDaysRequest) (map[string]any, error) {
	if err := validateWorkoutDays(req.Days); err != nil {
		return nil, err
	}
	templatePID, err := parsePublicID(templatePublicID)
	if err != nil {
		return nil, err
	}
	err = txtx.WithTx(ctx, s.repo.DB(), func(tx pgx.Tx) error {
		templateID, err := s.repo.WorkoutTemplateID(ctx, tx, auth.GymID, templatePID)
		if err != nil {
			return err
		}
		if err := s.repo.DeleteTemplateDays(ctx, tx, auth.GymID, templateID); err != nil {
			return err
		}
		for _, d := range req.Days {
			muscleID, err := s.resolveOptionalMuscleID(ctx, tx, auth.GymID, d.MuscleGroupPublicID, d.IsRestDay)
			if err != nil {
				return err
			}
			if err := s.repo.InsertTemplateDay(ctx, tx, token.GeneratePublicID(), auth.GymID, templateID, muscleID, d.DayOfWeek, d.Title, d.Description, d.IsRestDay); err != nil {
				return err
			}
		}
		return s.repo.LogAudit(ctx, tx, auth.GymID, &auth.UserID, "workout_template_days.updated", "workout_template_days", nil, req)
	})
	if err != nil {
		return nil, err
	}
	days, err := s.TemplateDays(ctx, auth, templatePublicID)
	if err != nil {
		return nil, err
	}
	return map[string]any{"days": days}, nil
}

func validateWorkoutDays(days []WorkoutDayRequest) error {
	if len(days) == 0 || len(days) > 7 {
		return apperror.BadRequest("Jumlah hari harus 1 sampai 7")
	}
	seen := map[int]bool{}
	for _, d := range days {
		if !day.ValidateDayOfWeek(d.DayOfWeek) {
			return apperror.BadRequest("day_of_week harus 1 sampai 7")
		}
		if seen[d.DayOfWeek] {
			return apperror.BadRequest("day_of_week tidak boleh duplicate")
		}
		seen[d.DayOfWeek] = true
		if !d.IsRestDay && (d.MuscleGroupPublicID == nil || *d.MuscleGroupPublicID == "") {
			return apperror.BadRequest("muscle_group_public_id wajib untuk hari latihan")
		}
	}
	return nil
}

func (s *Service) resolveOptionalMuscleID(ctx context.Context, db txtx.DBTX, gymID int64, raw *string, rest bool) (*int64, error) {
	if rest {
		return nil, nil
	}
	if raw == nil || *raw == "" {
		return nil, apperror.BadRequest("muscle_group_public_id wajib")
	}
	pid, err := parsePublicID(*raw)
	if err != nil {
		return nil, err
	}
	id, err := s.repo.ActiveMuscleGroupID(ctx, db, gymID, pid)
	if err != nil {
		return nil, apperror.BadRequest("Muscle group tidak valid")
	}
	return &id, nil
}

func (s *Service) AssignTemplateToMember(ctx context.Context, auth middleware.AuthUser, memberPublicID string, req AssignWorkoutTemplateRequest) (map[string]any, error) {
	memberPID, err := parsePublicID(memberPublicID)
	if err != nil {
		return nil, err
	}
	templatePID, err := parsePublicID(req.WorkoutTemplatePublicID)
	if err != nil {
		return nil, err
	}
	var schedulePublicID uuid.UUID
	err = txtx.WithTx(ctx, s.repo.DB(), func(tx pgx.Tx) error {
		memberID, memberName, err := s.repo.ActiveMemberForWorkout(ctx, tx, auth.GymID, memberPID)
		if err != nil {
			return err
		}
		templateID, templateName, err := s.repo.ActiveWorkoutTemplate(ctx, tx, auth.GymID, templatePID)
		if err != nil {
			return err
		}
		if err := s.repo.DeactivateActiveMemberSchedules(ctx, tx, auth.GymID, memberID); err != nil {
			return err
		}
		name := req.Name
		if name == "" {
			name = memberName + " - " + templateName
		}
		schedulePublicID = token.GeneratePublicID()
		scheduleID, err := s.repo.InsertMemberWorkoutSchedule(ctx, tx, schedulePublicID, auth.GymID, memberID, templateID, name, auth.UserID)
		if err != nil {
			return err
		}
		if err := s.repo.CopyTemplateDaysToMemberSchedule(ctx, tx, scheduleID, memberID, auth.GymID, templateID); err != nil {
			return err
		}
		return s.repo.LogAudit(ctx, tx, auth.GymID, &auth.UserID, "member_workout_schedule.assigned", "member_workout_schedules", &scheduleID, map[string]any{"member_id": memberID, "template_id": templateID})
	})
	if err != nil {
		return nil, err
	}
	return s.ActiveMemberSchedule(ctx, auth, memberPublicID)
}

func (s *Service) ActiveMemberSchedule(ctx context.Context, auth middleware.AuthUser, memberPublicID string) (map[string]any, error) {
	memberPID, err := parsePublicID(memberPublicID)
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.ActiveMemberSchedule(ctx, auth.GymID, memberPID)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, apperror.NotFound("Jadwal aktif tidak ditemukan")
	}
	days, err := s.scheduleDaysBySchedulePublicID(ctx, auth.GymID, rows[0]["public_id"].(uuid.UUID))
	if err != nil {
		return nil, err
	}
	rows[0]["days"] = days
	return rows[0], nil
}

func (s *Service) scheduleDaysBySchedulePublicID(ctx context.Context, gymID int64, schedulePublicID uuid.UUID) ([]map[string]any, error) {
	rows, err := s.repo.MemberScheduleDays(ctx, gymID, schedulePublicID)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		row["day_name"] = day.DayOfWeekName(intNum(row["day_of_week"]))
		if row["muscle_group_name"] != nil {
			row["muscle_group"] = map[string]any{"public_id": row["muscle_group_public_id"], "name": row["muscle_group_name"], "code": row["muscle_group_code"]}
		}
		delete(row, "muscle_group_public_id")
		delete(row, "muscle_group_name")
		delete(row, "muscle_group_code")
	}
	return rows, nil
}

func (s *Service) ReplaceMemberScheduleDays(ctx context.Context, auth middleware.AuthUser, memberPublicID string, schedulePublicID string, req ReplaceWorkoutDaysRequest) (map[string]any, error) {
	if err := validateWorkoutDays(req.Days); err != nil {
		return nil, err
	}
	memberPID, err := parsePublicID(memberPublicID)
	if err != nil {
		return nil, err
	}
	schedulePID, err := parsePublicID(schedulePublicID)
	if err != nil {
		return nil, err
	}
	err = txtx.WithTx(ctx, s.repo.DB(), func(tx pgx.Tx) error {
		memberID, scheduleID, err := s.repo.MemberAndScheduleID(ctx, tx, auth.GymID, memberPID, schedulePID)
		if err != nil {
			return err
		}
		if err := s.repo.DeleteMemberScheduleDays(ctx, tx, auth.GymID, scheduleID); err != nil {
			return err
		}
		for _, d := range req.Days {
			muscleID, err := s.resolveOptionalMuscleID(ctx, tx, auth.GymID, d.MuscleGroupPublicID, d.IsRestDay)
			if err != nil {
				return err
			}
			if err := s.repo.InsertMemberScheduleDay(ctx, tx, token.GeneratePublicID(), auth.GymID, scheduleID, memberID, muscleID, d.DayOfWeek, d.Title, d.Description, d.IsRestDay); err != nil {
				return err
			}
		}
		if err := s.repo.MarkMemberScheduleCustom(ctx, tx, auth.GymID, scheduleID); err != nil {
			return err
		}
		return s.repo.LogAudit(ctx, tx, auth.GymID, &auth.UserID, "member_workout_schedule.customized", "member_workout_schedules", &scheduleID, req)
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{"schedule_public_id": schedulePID, "is_custom": true}, nil
}

func (s *Service) PublicWorkoutSchedule(ctx context.Context, qrToken string) (map[string]any, error) {
	qr, err := s.resolveQR(ctx, qrToken)
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.PublicActiveSchedule(ctx, qr.GymID, qr.MemberID)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, apperror.NotFound("Jadwal aktif tidak ditemukan")
	}
	days, err := s.scheduleDaysBySchedulePublicID(ctx, qr.GymID, rows[0]["public_id"].(uuid.UUID))
	if err != nil {
		return nil, err
	}
	return map[string]any{"member_name": qr.MemberName, "schedule_name": rows[0]["name"], "days": days}, nil
}

func (s *Service) PublicTodayWorkoutSchedule(ctx context.Context, qrToken string) (map[string]any, error) {
	qr, err := s.resolveQR(ctx, qrToken)
	if err != nil {
		return nil, err
	}
	today := datetime.TodayInTimezone(qr.Timezone)
	dayNum := day.GetTodayDayOfWeekByTimezone(qr.Timezone)
	rows, err := s.repo.PublicTodaySchedule(ctx, qr.GymID, qr.MemberID, dayNum)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, apperror.NotFound("Jadwal hari ini tidak ditemukan")
	}
	row := rows[0]
	row["date"] = today.Format("2006-01-02")
	row["day_name"] = day.DayOfWeekName(dayNum)
	if row["muscle_group_name"] != nil {
		row["muscle_group"] = map[string]any{"name": row["muscle_group_name"], "code": row["muscle_group_code"]}
	}
	delete(row, "muscle_group_name")
	delete(row, "muscle_group_code")
	return row, nil
}

func (s *Service) StartWorkoutSession(ctx context.Context, req StartWorkoutSessionRequest) (map[string]any, error) {
	qr, err := s.resolveQR(ctx, req.QRToken)
	if err != nil {
		return nil, err
	}
	now := datetime.NowInTimezone(qr.Timezone)
	today := datetime.DateOnly(now)
	var scheduleDayID *int64
	var muscleID *int64
	title := strings.TrimSpace(req.Title)
	if req.ScheduleDayPublicID != "" {
		pid, err := parsePublicID(req.ScheduleDayPublicID)
		if err != nil {
			return nil, err
		}
		rows, err := s.repo.ScheduleDayForWorkout(ctx, qr.GymID, qr.MemberID, pid)
		if err != nil || len(rows) == 0 {
			return nil, apperror.BadRequest("Schedule day tidak valid")
		}
		id := rows[0]["id"].(int64)
		scheduleDayID = &id
		title = fmt.Sprint(rows[0]["title"])
		if rows[0]["muscle_group_id"] != nil {
			id := rows[0]["muscle_group_id"].(int64)
			muscleID = &id
		}
	} else {
		if title == "" {
			return nil, apperror.BadRequest("title wajib jika start manual")
		}
		if req.MuscleGroupPublicID != "" {
			mg := req.MuscleGroupPublicID
			muscleID, err = s.resolveOptionalMuscleID(ctx, s.repo.DB(), qr.GymID, &mg, false)
			if err != nil {
				return nil, err
			}
		}
	}
	checkinID, _ := s.repo.ValidCheckinIDForWorkout(ctx, qr.GymID, qr.MemberID, today)
	rows, err := s.repo.InsertWorkoutSessionStart(ctx, token.GeneratePublicID(), qr.GymID, qr.MemberID, checkinID, scheduleDayID, muscleID, today, title, now)
	if err != nil {
		return nil, err
	}
	return rows[0], nil
}

func (s *Service) FinishWorkoutSession(ctx context.Context, sessionPublicID string, req FinishWorkoutSessionRequest) (map[string]any, error) {
	qr, err := s.resolveQR(ctx, req.QRToken)
	if err != nil {
		return nil, err
	}
	sessionPID, err := parsePublicID(sessionPublicID)
	if err != nil {
		return nil, err
	}
	var out []map[string]any
	err = txtx.WithTx(ctx, s.repo.DB(), func(tx pgx.Tx) error {
		startedAt, err := s.repo.StartedWorkoutSession(ctx, tx, qr.GymID, qr.MemberID, sessionPID)
		if err != nil {
			return err
		}
		now := datetime.NowInTimezone(qr.Timezone)
		seconds := duration.CalculateDurationSeconds(startedAt, now)
		rows, err := s.repo.FinishWorkoutSession(ctx, tx, qr.GymID, qr.MemberID, sessionPID, now, seconds, req.Notes)
		if err != nil {
			return err
		}
		out = rows
		return s.repo.LogAudit(ctx, tx, qr.GymID, nil, "member_workout_session.completed", "member_workout_sessions", nil, out[0])
	})
	if err != nil {
		return nil, err
	}
	out[0]["duration_text"] = duration.SecondsToHuman(intNum(out[0]["duration_seconds"]))
	return out[0], nil
}

func (s *Service) WorkoutHistory(ctx context.Context, gymID, memberID int64, p pagination.Params) (PageResult, error) {
	total, err := s.repo.CountWorkoutSessions(ctx, gymID, memberID)
	if err != nil {
		return PageResult{}, err
	}
	items, err := s.repo.WorkoutHistory(ctx, gymID, memberID, p.Limit, p.Offset())
	if err != nil {
		return PageResult{}, err
	}
	for _, item := range items {
		item["duration_text"] = duration.SecondsToHuman(intNum(item["duration_seconds"]))
	}
	return PageResult{Items: items, Meta: pagination.NewMeta(p.Page, p.Limit, total)}, nil
}

func (s *Service) PublicWorkoutHistory(ctx context.Context, qrToken string, p pagination.Params) (PageResult, error) {
	qr, err := s.resolveQR(ctx, qrToken)
	if err != nil {
		return PageResult{}, err
	}
	return s.WorkoutHistory(ctx, qr.GymID, qr.MemberID, p)
}

func (s *Service) AdminMemberWorkoutHistory(ctx context.Context, auth middleware.AuthUser, memberPublicID string, p pagination.Params) (PageResult, error) {
	memberPID, err := parsePublicID(memberPublicID)
	if err != nil {
		return PageResult{}, err
	}
	memberID, err := s.repo.ResolveID(ctx, "members", auth.GymID, memberPID)
	if err != nil {
		return PageResult{}, err
	}
	return s.WorkoutHistory(ctx, auth.GymID, memberID, p)
}

func (s *Service) CreateManualWorkoutSession(ctx context.Context, auth middleware.AuthUser, memberPublicID string, req ManualWorkoutSessionRequest) (map[string]any, error) {
	memberPID, err := parsePublicID(memberPublicID)
	if err != nil {
		return nil, err
	}
	memberID, err := s.repo.ResolveID(ctx, "members", auth.GymID, memberPID)
	if err != nil {
		return nil, err
	}
	workoutDate, err := dateFromMap(map[string]any{"workout_date": req.WorkoutDate}, "workout_date", datetime.TodayInTimezone(s.cfg.DefaultTimezone))
	if err != nil {
		return nil, err
	}
	var muscleID *int64
	if req.MuscleGroupPublicID != "" {
		raw := req.MuscleGroupPublicID
		muscleID, err = s.resolveOptionalMuscleID(ctx, s.repo.DB(), auth.GymID, &raw, false)
		if err != nil {
			return nil, err
		}
	}
	rows, err := s.repo.InsertManualWorkoutSession(ctx, token.GeneratePublicID(), auth.GymID, memberID, muscleID, workoutDate, req.Title, req.Notes, req.DurationSeconds, adminWorkoutSource(auth.Role), auth.UserID)
	if err != nil {
		return nil, err
	}
	rows[0]["duration_text"] = duration.SecondsToHuman(req.DurationSeconds)
	return rows[0], nil
}

func (s *Service) listWeightLogs(ctx context.Context, gymID, memberID int64) (map[string]any, error) {
	items, err := s.repo.WeightLogs(ctx, gymID, memberID)
	if err != nil {
		return nil, err
	}
	summary := map[string]any{"first_weight": nil, "latest_weight": nil, "difference": nil}
	if len(items) > 0 {
		first := floatNum(items[0]["weight_kg"])
		latest := floatNum(items[len(items)-1]["weight_kg"])
		summary["first_weight"] = first
		summary["latest_weight"] = latest
		summary["difference"] = latest - first
	}
	return map[string]any{"data": items, "summary": summary}, nil
}

func floatNum(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int64:
		return float64(n)
	case int:
		return float64(n)
	case string:
		var f float64
		_, _ = fmt.Sscanf(n, "%f", &f)
		return f
	default:
		return 0
	}
}

func (s *Service) AdminWeightLogs(ctx context.Context, auth middleware.AuthUser, memberPublicID string) (map[string]any, error) {
	memberPID, err := parsePublicID(memberPublicID)
	if err != nil {
		return nil, err
	}
	memberID, err := s.repo.ResolveID(ctx, "members", auth.GymID, memberPID)
	if err != nil {
		return nil, err
	}
	return s.listWeightLogs(ctx, auth.GymID, memberID)
}

func (s *Service) CreateAdminWeightLog(ctx context.Context, auth middleware.AuthUser, memberPublicID string, req CreateWeightLogRequest) (map[string]any, error) {
	memberPID, err := parsePublicID(memberPublicID)
	if err != nil {
		return nil, err
	}
	memberID, err := s.repo.ResolveID(ctx, "members", auth.GymID, memberPID)
	if err != nil {
		return nil, err
	}
	measuredDate, err := dateFromMap(map[string]any{"measured_date": req.MeasuredDate}, "measured_date", datetime.TodayInTimezone(s.cfg.DefaultTimezone))
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.InsertAdminWeightLog(ctx, token.GeneratePublicID(), auth.GymID, memberID, measuredDate, req.WeightKG, req.Notes, adminWorkoutSource(auth.Role), auth.UserID)
	if err != nil {
		return nil, err
	}
	return rows[0], nil
}

func adminWorkoutSource(role string) string {
	if role == "trainer" {
		return "trainer"
	}
	return "admin"
}

func (s *Service) UpdateAdminWeightLog(ctx context.Context, auth middleware.AuthUser, memberPublicID, weightLogPublicID string, req CreateWeightLogRequest) (map[string]any, error) {
	memberPID, err := parsePublicID(memberPublicID)
	if err != nil {
		return nil, err
	}
	logPID, err := parsePublicID(weightLogPublicID)
	if err != nil {
		return nil, err
	}
	memberID, err := s.repo.ResolveID(ctx, "members", auth.GymID, memberPID)
	if err != nil {
		return nil, err
	}
	measuredDate, err := dateFromMap(map[string]any{"measured_date": req.MeasuredDate}, "measured_date", datetime.TodayInTimezone(s.cfg.DefaultTimezone))
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.UpdateAdminWeightLog(ctx, auth.GymID, memberID, logPID, measuredDate, req.WeightKG, req.Notes)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, apperror.NotFound("Weight log tidak ditemukan")
	}
	return rows[0], nil
}

func (s *Service) DeleteAdminWeightLog(ctx context.Context, auth middleware.AuthUser, memberPublicID, weightLogPublicID string) error {
	memberPID, err := parsePublicID(memberPublicID)
	if err != nil {
		return err
	}
	logPID, err := parsePublicID(weightLogPublicID)
	if err != nil {
		return err
	}
	memberID, err := s.repo.ResolveID(ctx, "members", auth.GymID, memberPID)
	if err != nil {
		return err
	}
	return s.repo.DeleteWeightLog(ctx, auth.GymID, memberID, logPID)
}

func (s *Service) PublicWeightLogs(ctx context.Context, qrToken string) (map[string]any, error) {
	qr, err := s.resolveQR(ctx, qrToken)
	if err != nil {
		return nil, err
	}
	return s.listWeightLogs(ctx, qr.GymID, qr.MemberID)
}

func (s *Service) CreatePublicWeightLog(ctx context.Context, req PublicCreateWeightLogRequest) (map[string]any, error) {
	qr, err := s.resolveQR(ctx, req.QRToken)
	if err != nil {
		return nil, err
	}
	measuredDate, err := dateFromMap(map[string]any{"measured_date": req.MeasuredDate}, "measured_date", datetime.TodayInTimezone(qr.Timezone))
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.InsertPublicWeightLog(ctx, token.GeneratePublicID(), qr.GymID, qr.MemberID, measuredDate, req.WeightKG, req.Notes)
	if err != nil {
		return nil, err
	}
	return rows[0], nil
}

func (s *Service) MemberProgressReport(ctx context.Context, auth middleware.AuthUser, memberPublicID string) (map[string]any, error) {
	memberPID, err := parsePublicID(memberPublicID)
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.MemberProgressReport(ctx, auth.GymID, memberPID)
	if err != nil || len(rows) == 0 {
		return nil, apperror.NotFound("Member tidak ditemukan")
	}
	row := rows[0]
	row["total_duration_text"] = duration.SecondsToHuman(intNum(row["total_duration_seconds"]))
	row["average_duration_text"] = duration.SecondsToHuman(intNum(row["average_duration_seconds"]))
	row["weight_difference"] = floatNum(row["latest_weight"]) - floatNum(row["first_weight"])
	return row, nil
}
