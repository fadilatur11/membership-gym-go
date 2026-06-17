package member

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"membership-gym/config"
	"membership-gym/internal/middleware"
	"membership-gym/pkg/datetime"
	apperror "membership-gym/pkg/errors"
	"membership-gym/pkg/pagination"
	"membership-gym/pkg/sqlutil"
	"membership-gym/pkg/token"
	txtx "membership-gym/pkg/tx"
)

type PageResult struct {
	Items []map[string]any
	Meta  pagination.Meta
}

type Service struct {
	cfg  config.Config
	repo *Repository
}

func NewService(cfg config.Config, repo *Repository) *Service {
	return &Service{cfg: cfg, repo: repo}
}

func (s *Service) PlanAllows(ctx context.Context, auth middleware.AuthUser, feature string) (bool, error) {
	if feature == "" {
		return true, nil
	}
	rows, err := s.repo.ActivePlanFeatures(ctx, auth.GymID)
	if err != nil {
		return false, err
	}
	if len(rows) == 0 {
		return false, nil
	}
	if fmt.Sprint(rows[0]["code"]) == "pro" {
		return true, nil
	}
	features, _ := rows[0]["features"].([]any)
	for _, item := range features {
		value := fmt.Sprint(item)
		if value == "all" || value == feature {
			return true, nil
		}
	}
	return false, nil
}

func (s *Service) ListMembers(ctx context.Context, auth middleware.AuthUser, p pagination.Params, q map[string]string) (PageResult, error) {
	where := []string{"gym_id=$1"}
	args := []any{auth.GymID}
	if search := strings.TrimSpace(q["search"]); search != "" {
		args = append(args, sqlutil.LikeSearch(search))
		where = append(where, fmt.Sprintf("(LOWER(full_name) LIKE $%d OR LOWER(member_code) LIKE $%d OR phone LIKE $%d)", len(args), len(args), len(args)))
	}
	if status := strings.TrimSpace(q["status"]); status != "" {
		args = append(args, status)
		where = append(where, fmt.Sprintf("status=$%d", len(args)))
	}
	items, total, err := s.repo.ListMembers(ctx, where, args, p.Limit, p.Offset())
	if err != nil {
		return PageResult{}, err
	}
	return PageResult{Items: items, Meta: pagination.NewMeta(p.Page, p.Limit, total)}, nil
}

func (s *Service) DetailMember(ctx context.Context, auth middleware.AuthUser, publicID string) (map[string]any, error) {
	id, err := parsePublicID(publicID)
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.DetailMember(ctx, auth.GymID, id)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, apperror.NotFound("Data tidak ditemukan")
	}
	return rows[0], nil
}

func (s *Service) CreateMember(ctx context.Context, auth middleware.AuthUser, req CreateMemberRequest) (map[string]any, error) {
	joinedAt, err := dateFromString(req.JoinedAt, datetime.TodayInTimezone(s.cfg.DefaultTimezone), "joined_at")
	if err != nil {
		return nil, err
	}
	birthDate, err := optionalDate(req.BirthDate, "birth_date")
	if err != nil {
		return nil, err
	}
	var rows []map[string]any
	err = txtx.WithTx(ctx, s.repo.DB(), func(tx pgx.Tx) error {
		var e error
		rows, e = s.repo.InsertMember(ctx, tx, token.GeneratePublicID(), auth.GymID, req.MemberCode, req.FullName, req.Phone, req.Email, req.Gender, birthDate, req.Address, req.EmergencyContactName, req.EmergencyContactPhone, joinedAt, req.Notes)
		if e != nil {
			return e
		}
		return s.repo.LogAudit(ctx, tx, auth.GymID, &auth.UserID, "member.created", "members", nil, rows[0])
	})
	if err != nil {
		return nil, err
	}
	return rows[0], nil
}

func (s *Service) UpdateMember(ctx context.Context, auth middleware.AuthUser, publicID string, req UpdateMemberRequest) (map[string]any, error) {
	id, err := parsePublicID(publicID)
	if err != nil {
		return nil, err
	}
	sets, args := []string{}, []any{auth.GymID, id}
	fields := map[string]any{
		"full_name": req.FullName, "phone": req.Phone, "email": req.Email, "gender": req.Gender,
		"address": req.Address, "status": req.Status, "notes": req.Notes,
		"emergency_contact_name": req.EmergencyContactName, "emergency_contact_phone": req.EmergencyContactPhone,
	}
	for col, value := range fields {
		if strings.TrimSpace(fmt.Sprint(value)) != "" {
			args = append(args, value)
			sets = append(sets, fmt.Sprintf("%s=$%d", col, len(args)))
		}
	}
	if len(sets) == 0 {
		return s.DetailMember(ctx, auth, publicID)
	}
	sets = append(sets, "updated_at=NOW()")
	var rows []map[string]any
	err = txtx.WithTx(ctx, s.repo.DB(), func(tx pgx.Tx) error {
		var e error
		rows, e = s.repo.UpdateMember(ctx, tx, args, sets)
		if e != nil {
			return e
		}
		if len(rows) == 0 {
			return apperror.NotFound("Data tidak ditemukan")
		}
		return s.repo.LogAudit(ctx, tx, auth.GymID, &auth.UserID, "members.updated", "members", nil, rows[0])
	})
	if err != nil {
		return nil, err
	}
	return rows[0], nil
}

func (s *Service) DeactivateMember(ctx context.Context, auth middleware.AuthUser, publicID string) (map[string]any, error) {
	id, err := parsePublicID(publicID)
	if err != nil {
		return nil, err
	}
	var rows []map[string]any
	err = txtx.WithTx(ctx, s.repo.DB(), func(tx pgx.Tx) error {
		var e error
		rows, e = s.repo.ChangeMemberStatus(ctx, tx, auth.GymID, id, "inactive")
		if e != nil {
			return e
		}
		if len(rows) == 0 {
			return apperror.NotFound("Data tidak ditemukan")
		}
		return s.repo.LogAudit(ctx, tx, auth.GymID, &auth.UserID, "member.deactivated", "members", nil, rows[0])
	})
	if err != nil {
		return nil, err
	}
	return rows[0], nil
}

func (s *Service) PublicIdentity(ctx context.Context, qrToken string) (map[string]any, error) {
	rows, err := s.repo.PublicQRIdentity(ctx, qrToken)
	if err != nil || len(rows) == 0 {
		return nil, apperror.NotFound("QR tidak valid")
	}
	delete(rows[0], "member_id")
	delete(rows[0], "gym_id")
	return rows[0], nil
}

func (s *Service) PublicStatus(ctx context.Context, qrToken string) (map[string]any, error) {
	rows, err := s.repo.PublicMemberStatus(ctx, qrToken)
	if err != nil || len(rows) == 0 {
		return nil, apperror.NotFound("QR tidak valid")
	}
	if end, ok := rows[0]["subscription_end_date"].(string); ok && end != "" {
		t, _ := time.Parse("2006-01-02", end)
		rows[0]["days_remaining"] = datetime.DaysRemaining(t, datetime.NowInTimezone(fmt.Sprint(rows[0]["timezone"])))
	}
	delete(rows[0], "timezone")
	return rows[0], nil
}

func dateFromString(raw string, fallback time.Time, field string) (time.Time, error) {
	if raw == "" {
		return fallback, nil
	}
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return time.Time{}, apperror.BadRequest(field + " tidak valid")
	}
	return t, nil
}

func optionalDate(raw string, field string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return nil, apperror.BadRequest(field + " tidak valid")
	}
	return &t, nil
}

func parsePublicID(raw string) (uuid.UUID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, apperror.BadRequest("public_id tidak valid")
	}
	return id, nil
}
