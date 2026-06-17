package reminder_rule

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"membership-gym/internal/middleware"
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
	repo *Repository
}

func NewService(repo *Repository) *Service { return &Service{repo: repo} }

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

func (s *Service) ListRules(ctx context.Context, auth middleware.AuthUser, p pagination.Params, q map[string]string) (PageResult, error) {
	where := []string{"gym_id=$1"}
	args := []any{auth.GymID}
	if search := strings.TrimSpace(q["search"]); search != "" {
		args = append(args, sqlutil.LikeSearch(search))
		where = append(where, fmt.Sprintf("LOWER(name) LIKE $%d", len(args)))
	}
	if active := strings.TrimSpace(q["is_active"]); active != "" {
		args = append(args, active)
		where = append(where, fmt.Sprintf("is_active=$%d", len(args)))
	}
	items, total, err := s.repo.ListRules(ctx, where, args, p.Limit, p.Offset())
	if err != nil {
		return PageResult{}, err
	}
	return PageResult{Items: items, Meta: pagination.NewMeta(p.Page, p.Limit, total)}, nil
}

func (s *Service) DetailRule(ctx context.Context, auth middleware.AuthUser, publicID string) (map[string]any, error) {
	id, err := parsePublicID(publicID)
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.DetailRule(ctx, auth.GymID, id)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, apperror.NotFound("Data tidak ditemukan")
	}
	return rows[0], nil
}

func (s *Service) CreateRule(ctx context.Context, auth middleware.AuthUser, req CreateReminderRuleRequest) (map[string]any, error) {
	if req.DaysBeforeExpiry < 0 {
		return nil, apperror.BadRequest("days_before_expiry tidak valid")
	}
	channel := req.Channel
	if channel == "" {
		channel = "whatsapp"
	}
	var rows []map[string]any
	err := txtx.WithTx(ctx, s.repo.DB(), func(tx pgx.Tx) error {
		var e error
		rows, e = s.repo.InsertRule(ctx, tx, token.GeneratePublicID(), auth.GymID, req.Name, req.DaysBeforeExpiry, channel, req.MessageTemplate)
		if e != nil {
			return e
		}
		return s.repo.LogAudit(ctx, tx, auth.GymID, &auth.UserID, "reminder_rule.created", "reminder_rules", nil, rows[0])
	})
	if err != nil {
		return nil, err
	}
	return rows[0], nil
}

func (s *Service) UpdateRule(ctx context.Context, auth middleware.AuthUser, publicID string, req UpdateReminderRuleRequest) (map[string]any, error) {
	id, err := parsePublicID(publicID)
	if err != nil {
		return nil, err
	}
	sets, args := []string{}, []any{auth.GymID, id}
	if req.Name != "" {
		args = append(args, req.Name)
		sets = append(sets, fmt.Sprintf("name=$%d", len(args)))
	}
	if req.DaysBeforeExpiry > 0 {
		args = append(args, req.DaysBeforeExpiry)
		sets = append(sets, fmt.Sprintf("days_before_expiry=$%d", len(args)))
	}
	if req.Channel != "" {
		args = append(args, req.Channel)
		sets = append(sets, fmt.Sprintf("channel=$%d", len(args)))
	}
	if req.MessageTemplate != "" {
		args = append(args, req.MessageTemplate)
		sets = append(sets, fmt.Sprintf("message_template=$%d", len(args)))
	}
	if req.IsActive != nil {
		args = append(args, *req.IsActive)
		sets = append(sets, fmt.Sprintf("is_active=$%d", len(args)))
	}
	if len(sets) == 0 {
		return s.DetailRule(ctx, auth, publicID)
	}
	sets = append(sets, "updated_at=NOW()")
	var rows []map[string]any
	err = txtx.WithTx(ctx, s.repo.DB(), func(tx pgx.Tx) error {
		var e error
		rows, e = s.repo.UpdateRule(ctx, tx, args, sets)
		if e != nil {
			return e
		}
		if len(rows) == 0 {
			return apperror.NotFound("Data tidak ditemukan")
		}
		return s.repo.LogAudit(ctx, tx, auth.GymID, &auth.UserID, "reminder_rules.updated", "reminder_rules", nil, rows[0])
	})
	if err != nil {
		return nil, err
	}
	return rows[0], nil
}

func (s *Service) DeactivateRule(ctx context.Context, auth middleware.AuthUser, publicID string) (map[string]any, error) {
	id, err := parsePublicID(publicID)
	if err != nil {
		return nil, err
	}
	var rows []map[string]any
	err = txtx.WithTx(ctx, s.repo.DB(), func(tx pgx.Tx) error {
		var e error
		rows, e = s.repo.ChangeRuleStatus(ctx, tx, auth.GymID, id, false)
		if e != nil {
			return e
		}
		if len(rows) == 0 {
			return apperror.NotFound("Data tidak ditemukan")
		}
		return s.repo.LogAudit(ctx, tx, auth.GymID, &auth.UserID, "reminder_rule.deactivated", "reminder_rules", nil, rows[0])
	})
	if err != nil {
		return nil, err
	}
	return rows[0], nil
}

func parsePublicID(raw string) (uuid.UUID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, apperror.BadRequest("public_id tidak valid")
	}
	return id, nil
}
