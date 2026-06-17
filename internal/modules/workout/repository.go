package workout

import (
	"context"
	"time"

	"github.com/google/uuid"
	"membership-gym/internal/models"
	txtx "membership-gym/pkg/tx"
)

func (r *Repository) V2ResolveQR(ctx context.Context, qrToken string) ([]map[string]any, error) {
	return r.QueryMaps(ctx, models.QueryV2ResolveQR, qrToken)
}

func (r *Repository) TemplateDays(ctx context.Context, gymID int64, templatePID uuid.UUID) ([]map[string]any, error) {
	return r.QueryMaps(ctx, models.QueryTemplateDays, gymID, templatePID)
}

func (r *Repository) WorkoutTemplateID(ctx context.Context, db txtx.DBTX, gymID int64, templatePID uuid.UUID) (int64, error) {
	var id int64
	err := db.QueryRow(ctx, models.QueryWorkoutTemplateID, gymID, templatePID).Scan(&id)
	return id, err
}

func (r *Repository) DeleteTemplateDays(ctx context.Context, db txtx.DBTX, gymID, templateID int64) error {
	_, err := db.Exec(ctx, models.QueryDeleteTemplateDays, gymID, templateID)
	return err
}

func (r *Repository) InsertTemplateDay(ctx context.Context, db txtx.DBTX, args ...any) error {
	_, err := db.Exec(ctx, models.QueryInsertTemplateDay, args...)
	return err
}

func (r *Repository) ActiveMuscleGroupID(ctx context.Context, db txtx.DBTX, gymID int64, publicID uuid.UUID) (int64, error) {
	var id int64
	err := db.QueryRow(ctx, models.QueryActiveMuscleGroupID, gymID, publicID).Scan(&id)
	return id, err
}

func (r *Repository) ActiveMemberForWorkout(ctx context.Context, db txtx.DBTX, gymID int64, memberPID uuid.UUID) (int64, string, error) {
	var id int64
	var name string
	err := db.QueryRow(ctx, models.QueryActiveMemberForWorkout, gymID, memberPID).Scan(&id, &name)
	return id, name, err
}

func (r *Repository) ActiveWorkoutTemplate(ctx context.Context, db txtx.DBTX, gymID int64, templatePID uuid.UUID) (int64, string, error) {
	var id int64
	var name string
	err := db.QueryRow(ctx, models.QueryActiveWorkoutTemplate, gymID, templatePID).Scan(&id, &name)
	return id, name, err
}

func (r *Repository) DeactivateActiveMemberSchedules(ctx context.Context, db txtx.DBTX, gymID, memberID int64) error {
	_, err := db.Exec(ctx, models.QueryDeactivateActiveMemberSchedules, gymID, memberID)
	return err
}

func (r *Repository) InsertMemberWorkoutSchedule(ctx context.Context, db txtx.DBTX, args ...any) (int64, error) {
	var id int64
	err := db.QueryRow(ctx, models.QueryInsertMemberWorkoutSchedule, args...).Scan(&id)
	return id, err
}

func (r *Repository) CopyTemplateDaysToMemberSchedule(ctx context.Context, db txtx.DBTX, scheduleID, memberID, gymID, templateID int64) error {
	_, err := db.Exec(ctx, models.QueryCopyTemplateDaysToMemberSchedule, scheduleID, memberID, gymID, templateID)
	return err
}

func (r *Repository) ActiveMemberSchedule(ctx context.Context, gymID int64, memberPID uuid.UUID) ([]map[string]any, error) {
	return r.QueryMaps(ctx, models.QueryActiveMemberSchedule, gymID, memberPID)
}

func (r *Repository) MemberScheduleDays(ctx context.Context, gymID int64, schedulePID uuid.UUID) ([]map[string]any, error) {
	return r.QueryMaps(ctx, models.QueryMemberScheduleDays, gymID, schedulePID)
}

func (r *Repository) MemberAndScheduleID(ctx context.Context, db txtx.DBTX, gymID int64, memberPID, schedulePID uuid.UUID) (int64, int64, error) {
	var memberID, scheduleID int64
	err := db.QueryRow(ctx, models.QueryMemberAndScheduleID, gymID, memberPID, schedulePID).Scan(&memberID, &scheduleID)
	return memberID, scheduleID, err
}

func (r *Repository) DeleteMemberScheduleDays(ctx context.Context, db txtx.DBTX, gymID, scheduleID int64) error {
	_, err := db.Exec(ctx, models.QueryDeleteMemberScheduleDays, gymID, scheduleID)
	return err
}

func (r *Repository) InsertMemberScheduleDay(ctx context.Context, db txtx.DBTX, args ...any) error {
	_, err := db.Exec(ctx, models.QueryInsertMemberScheduleDay, args...)
	return err
}

func (r *Repository) MarkMemberScheduleCustom(ctx context.Context, db txtx.DBTX, gymID, scheduleID int64) error {
	_, err := db.Exec(ctx, models.QueryMarkMemberScheduleCustom, gymID, scheduleID)
	return err
}

func (r *Repository) PublicActiveSchedule(ctx context.Context, gymID, memberID int64) ([]map[string]any, error) {
	return r.QueryMaps(ctx, models.QueryPublicActiveSchedule, gymID, memberID)
}

func (r *Repository) PublicTodaySchedule(ctx context.Context, gymID, memberID int64, dayOfWeek int) ([]map[string]any, error) {
	return r.QueryMaps(ctx, models.QueryPublicTodaySchedule, gymID, memberID, dayOfWeek)
}

func (r *Repository) ScheduleDayForWorkout(ctx context.Context, gymID, memberID int64, scheduleDayPID uuid.UUID) ([]map[string]any, error) {
	return r.QueryMaps(ctx, models.QueryScheduleDayForWorkout, gymID, memberID, scheduleDayPID)
}

func (r *Repository) ValidCheckinIDForWorkout(ctx context.Context, gymID, memberID int64, date time.Time) (*int64, error) {
	var id *int64
	err := r.db.QueryRow(ctx, models.QueryValidCheckinIDForWorkout, gymID, memberID, date).Scan(&id)
	return id, err
}

func (r *Repository) InsertWorkoutSessionStart(ctx context.Context, args ...any) ([]map[string]any, error) {
	return r.QueryMaps(ctx, models.QueryInsertWorkoutSessionStart, args...)
}

func (r *Repository) StartedWorkoutSession(ctx context.Context, db txtx.DBTX, gymID, memberID int64, sessionPID uuid.UUID) (time.Time, error) {
	var startedAt time.Time
	err := db.QueryRow(ctx, models.QueryStartedWorkoutSession, gymID, memberID, sessionPID).Scan(&startedAt)
	return startedAt, err
}

func (r *Repository) FinishWorkoutSession(ctx context.Context, db txtx.DBTX, args ...any) ([]map[string]any, error) {
	return r.QueryMapsWith(ctx, db, models.QueryFinishWorkoutSession, args...)
}

func (r *Repository) CountWorkoutSessions(ctx context.Context, gymID, memberID int64) (int64, error) {
	return r.Count(ctx, models.QueryCountWorkoutSessions, gymID, memberID)
}

func (r *Repository) WorkoutHistory(ctx context.Context, gymID, memberID int64, limit, offset int) ([]map[string]any, error) {
	return r.QueryMaps(ctx, models.QueryWorkoutHistory, gymID, memberID, limit, offset)
}

func (r *Repository) InsertManualWorkoutSession(ctx context.Context, args ...any) ([]map[string]any, error) {
	return r.QueryMaps(ctx, models.QueryInsertManualWorkoutSession, args...)
}

func (r *Repository) WeightLogs(ctx context.Context, gymID, memberID int64) ([]map[string]any, error) {
	return r.QueryMaps(ctx, models.QueryWeightLogs, gymID, memberID)
}

func (r *Repository) InsertAdminWeightLog(ctx context.Context, args ...any) ([]map[string]any, error) {
	return r.QueryMaps(ctx, models.QueryInsertAdminWeightLog, args...)
}

func (r *Repository) UpdateAdminWeightLog(ctx context.Context, args ...any) ([]map[string]any, error) {
	return r.QueryMaps(ctx, models.QueryUpdateAdminWeightLog, args...)
}

func (r *Repository) DeleteWeightLog(ctx context.Context, gymID, memberID int64, logPID uuid.UUID) error {
	return r.Exec(ctx, models.QueryDeleteWeightLog, gymID, memberID, logPID)
}

func (r *Repository) InsertPublicWeightLog(ctx context.Context, args ...any) ([]map[string]any, error) {
	return r.QueryMaps(ctx, models.QueryInsertPublicWeightLog, args...)
}

func (r *Repository) MemberProgressReport(ctx context.Context, gymID int64, memberPID uuid.UUID) ([]map[string]any, error) {
	return r.QueryMaps(ctx, models.QueryMemberProgressReport, gymID, memberPID)
}
