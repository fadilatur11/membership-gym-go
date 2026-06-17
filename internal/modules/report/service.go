package report

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"membership-gym/internal/middleware"
	"membership-gym/internal/modules/shared"
	apperror "membership-gym/pkg/errors"
	"membership-gym/pkg/pagination"
	"membership-gym/pkg/sqlutil"
)

var (
	paymentConfig = shared.ResourceConfig{
		Table: "payments", Select: "public_id, invoice_no, payment_type, payment_method, package_price, discount_amount, final_amount, status, paid_at, notes, created_at, updated_at",
		Filters: map[string]string{"status": "status", "payment_method": "payment_method"},
	}
	expenseConfig = shared.ResourceConfig{
		Table: "expenses", Select: "public_id, expense_category_id, title, description, amount, expense_date, status, created_at, updated_at",
		SearchSQL: "LOWER(title) LIKE $2", Filters: map[string]string{"status": "status"},
	}
	checkinConfig = shared.ResourceConfig{
		Table: "member_checkins", Select: "public_id, member_id, subscription_id, checkin_date, checkin_at, source, status, notes, created_at, updated_at",
		Filters: map[string]string{"status": "status", "source": "source"},
	}
	subscriptionConfig = shared.ResourceConfig{
		Table: "subscriptions", Select: "public_id, member_id, membership_package_id, start_date, end_date, status, source, created_at, updated_at",
		Filters: map[string]string{"status": "status"},
	}
	workoutSessionConfig = shared.ResourceConfig{
		Table: "member_workout_sessions", Select: "public_id, member_id, workout_date, title, notes, started_at, ended_at, duration_seconds, status, source, created_at, updated_at",
		SearchSQL: "LOWER(title) LIKE $2", Filters: map[string]string{"status": "status", "source": "source"},
	}
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

func (s *Service) Dashboard(ctx context.Context, auth middleware.AuthUser) (map[string]any, error) {
	rows, err := s.repo.Dashboard(ctx, auth.GymID)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return map[string]any{}, nil
	}
	out := rows[0]
	out["net_profit_this_month"] = int64Num(out["income_this_month"]) - int64Num(out["expense_this_month"])
	return out, nil
}

func (s *Service) ReportPayments(ctx context.Context, auth middleware.AuthUser, p pagination.Params, q map[string]string) (PageResult, error) {
	return s.list(ctx, auth, paymentConfig, p, q)
}

func (s *Service) ReportExpenses(ctx context.Context, auth middleware.AuthUser, p pagination.Params, q map[string]string) (PageResult, error) {
	return s.list(ctx, auth, expenseConfig, p, q)
}

func (s *Service) ReportCheckins(ctx context.Context, auth middleware.AuthUser, p pagination.Params, q map[string]string) (PageResult, error) {
	return s.list(ctx, auth, checkinConfig, p, q)
}

func (s *Service) ReportExpiredMembers(ctx context.Context, auth middleware.AuthUser, p pagination.Params, q map[string]string) (PageResult, error) {
	q["status"] = "expired"
	return s.list(ctx, auth, subscriptionConfig, p, q)
}

func (s *Service) WorkoutConsistency(ctx context.Context, auth middleware.AuthUser, p pagination.Params, q map[string]string) (PageResult, error) {
	return s.list(ctx, auth, workoutSessionConfig, p, q)
}

func (s *Service) MemberProgressReport(ctx context.Context, auth middleware.AuthUser, memberPublicID string) (map[string]any, error) {
	memberPID, err := uuid.Parse(memberPublicID)
	if err != nil {
		return nil, apperror.BadRequest("public_id tidak valid")
	}
	rows, err := s.repo.MemberProgressReport(ctx, auth.GymID, memberPID)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, apperror.NotFound("Member tidak ditemukan")
	}
	return rows[0], nil
}

func (s *Service) list(ctx context.Context, auth middleware.AuthUser, cfg shared.ResourceConfig, p pagination.Params, q map[string]string) (PageResult, error) {
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
	items, total, err := s.repo.List(ctx, cfg, where, args, p.Limit, p.Offset())
	if err != nil {
		return PageResult{}, err
	}
	return PageResult{Items: items, Meta: pagination.NewMeta(p.Page, p.Limit, total)}, nil
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
