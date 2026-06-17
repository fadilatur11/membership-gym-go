package workout

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"membership-gym/config"
	"membership-gym/internal/middleware"
	"membership-gym/internal/modules/shared"
	"membership-gym/pkg/datetime"
	apperror "membership-gym/pkg/errors"
	"membership-gym/pkg/pagination"
	"membership-gym/pkg/sqlutil"
	txtx "membership-gym/pkg/tx"
)

type Service struct {
	cfg  config.Config
	repo *Repository
}

type PageResult = shared.PageResult

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

func (s *Service) List(ctx context.Context, auth middleware.AuthUser, cfg shared.ResourceConfig, p pagination.Params, q map[string]string) (shared.PageResult, error) {
	where := []string{"gym_id=$1"}
	args := []any{auth.GymID}
	if search := strings.TrimSpace(q["search"]); search != "" && cfg.SearchSQL != "" {
		args = append(args, sqlutil.LikeSearch(search))
		where = append(where, strings.ReplaceAll(cfg.SearchSQL, "$2", fmt.Sprintf("$%d", len(args))))
	}
	for key, col := range cfg.Filters {
		if value := strings.TrimSpace(q[key]); value != "" {
			args = append(args, value)
			where = append(where, fmt.Sprintf("%s=$%d", col, len(args)))
		}
	}
	items, total, err := s.repo.ListGeneric(ctx, cfg, where, args, p.Limit, p.Offset())
	if err != nil {
		return shared.PageResult{}, err
	}
	return shared.PageResult{Items: items, Meta: pagination.NewMeta(p.Page, p.Limit, total)}, nil
}

func (s *Service) Detail(ctx context.Context, auth middleware.AuthUser, cfg shared.ResourceConfig, publicID string) (map[string]any, error) {
	id, err := parsePublicID(publicID)
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.DetailGeneric(ctx, cfg, auth.GymID, id)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, apperror.NotFound("Data tidak ditemukan")
	}
	return rows[0], nil
}

func (s *Service) UpdateGeneric(ctx context.Context, auth middleware.AuthUser, cfg shared.ResourceConfig, publicID string, req map[string]any) (map[string]any, error) {
	id, err := parsePublicID(publicID)
	if err != nil {
		return nil, err
	}
	sets, args := []string{}, []any{auth.GymID, id}
	for key, col := range cfg.UpdateFields {
		if value, ok := req[key]; ok {
			args = append(args, value)
			sets = append(sets, fmt.Sprintf("%s=$%d", col, len(args)))
		}
	}
	if len(sets) == 0 {
		return s.Detail(ctx, auth, cfg, publicID)
	}
	sets = append(sets, "updated_at=NOW()")
	var rows []map[string]any
	err = txtx.WithTx(ctx, s.repo.DB(), func(tx pgx.Tx) error {
		var e error
		rows, e = s.repo.UpdateGeneric(ctx, tx, cfg, args, sets)
		if e != nil {
			return e
		}
		if len(rows) == 0 {
			return apperror.NotFound("Data tidak ditemukan")
		}
		return s.repo.LogAudit(ctx, tx, auth.GymID, &auth.UserID, cfg.Table+".updated", cfg.Table, nil, rows[0])
	})
	if err != nil {
		return nil, err
	}
	return rows[0], nil
}

func (s *Service) ChangeStatus(ctx context.Context, auth middleware.AuthUser, cfg shared.ResourceConfig, publicID string, status any, action string) (map[string]any, error) {
	id, err := parsePublicID(publicID)
	if err != nil {
		return nil, err
	}
	var rows []map[string]any
	err = txtx.WithTx(ctx, s.repo.DB(), func(tx pgx.Tx) error {
		var e error
		rows, e = s.repo.ChangeStatusGeneric(ctx, tx, cfg, auth.GymID, id, status)
		if e != nil {
			return e
		}
		if len(rows) == 0 {
			return apperror.NotFound("Data tidak ditemukan")
		}
		return s.repo.LogAudit(ctx, tx, auth.GymID, &auth.UserID, action, cfg.Table, nil, rows[0])
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

func dateFromMap(req map[string]any, key string, fallback time.Time) (time.Time, error) {
	raw, _ := req[key].(string)
	if raw == "" {
		return fallback, nil
	}
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return time.Time{}, apperror.BadRequest(key + " tidak valid")
	}
	return t, nil
}

func intNum(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int32:
		return int(n)
	case int64:
		return int(n)
	default:
		return 0
	}
}

func int64Num(v any) int64 {
	switch n := v.(type) {
	case float64:
		return int64(n)
	case int:
		return int64(n)
	case int64:
		return n
	default:
		return 0
	}
}

func today(cfg config.Config) time.Time {
	return datetime.TodayInTimezone(cfg.DefaultTimezone)
}
