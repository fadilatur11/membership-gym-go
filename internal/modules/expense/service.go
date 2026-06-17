package expense

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

func NewService(cfg config.Config, repo *Repository) *Service { return &Service{cfg: cfg, repo: repo} }

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

func (s *Service) ListExpenses(ctx context.Context, auth middleware.AuthUser, p pagination.Params, q map[string]string) (PageResult, error) {
	where := []string{"gym_id=$1"}
	args := []any{auth.GymID}
	if search := strings.TrimSpace(q["search"]); search != "" {
		args = append(args, sqlutil.LikeSearch(search))
		where = append(where, fmt.Sprintf("LOWER(title) LIKE $%d", len(args)))
	}
	if status := strings.TrimSpace(q["status"]); status != "" {
		args = append(args, status)
		where = append(where, fmt.Sprintf("status=$%d", len(args)))
	}
	items, total, err := s.repo.ListExpenses(ctx, where, args, p.Limit, p.Offset())
	if err != nil {
		return PageResult{}, err
	}
	return PageResult{Items: items, Meta: pagination.NewMeta(p.Page, p.Limit, total)}, nil
}

func (s *Service) DetailExpense(ctx context.Context, auth middleware.AuthUser, publicID string) (map[string]any, error) {
	id, err := parsePublicID(publicID)
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.DetailExpense(ctx, auth.GymID, id)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, apperror.NotFound("Data tidak ditemukan")
	}
	return rows[0], nil
}

func (s *Service) CreateExpense(ctx context.Context, auth middleware.AuthUser, req CreateExpenseRequest) (map[string]any, error) {
	categoryPID, err := parsePublicID(req.ExpenseCategoryPublicID)
	if err != nil {
		return nil, err
	}
	categoryID, err := s.repo.ResolveID(ctx, "expense_categories", auth.GymID, categoryPID)
	if err != nil {
		return nil, apperror.BadRequest("Kategori expense tidak valid")
	}
	expenseDate, err := dateFromString(req.ExpenseDate, datetime.TodayInTimezone(s.cfg.DefaultTimezone), "expense_date")
	if err != nil {
		return nil, err
	}
	status := req.Status
	if status == "" {
		status = "approved"
	}
	var rows []map[string]any
	err = txtx.WithTx(ctx, s.repo.DB(), func(tx pgx.Tx) error {
		var e error
		rows, e = s.repo.InsertExpense(ctx, tx, token.GeneratePublicID(), auth.GymID, categoryID, req.Title, req.Description, req.Amount, expenseDate, status, auth.UserID)
		if e != nil {
			return e
		}
		return s.repo.LogAudit(ctx, tx, auth.GymID, &auth.UserID, "expense.created", "expenses", nil, rows[0])
	})
	if err != nil {
		return nil, err
	}
	return rows[0], nil
}

func (s *Service) UpdateExpense(ctx context.Context, auth middleware.AuthUser, publicID string, req UpdateExpenseRequest) (map[string]any, error) {
	id, err := parsePublicID(publicID)
	if err != nil {
		return nil, err
	}
	sets, args := []string{}, []any{auth.GymID, id}
	if req.Title != "" {
		args = append(args, req.Title)
		sets = append(sets, fmt.Sprintf("title=$%d", len(args)))
	}
	if req.Description != "" {
		args = append(args, req.Description)
		sets = append(sets, fmt.Sprintf("description=$%d", len(args)))
	}
	if req.Amount > 0 {
		args = append(args, req.Amount)
		sets = append(sets, fmt.Sprintf("amount=$%d", len(args)))
	}
	if req.ExpenseDate != "" {
		expenseDate, e := dateFromString(req.ExpenseDate, time.Time{}, "expense_date")
		if e != nil {
			return nil, e
		}
		args = append(args, expenseDate)
		sets = append(sets, fmt.Sprintf("expense_date=$%d", len(args)))
	}
	if req.Status != "" {
		args = append(args, req.Status)
		sets = append(sets, fmt.Sprintf("status=$%d", len(args)))
	}
	if len(sets) == 0 {
		return s.DetailExpense(ctx, auth, publicID)
	}
	sets = append(sets, "updated_at=NOW()")
	var rows []map[string]any
	err = txtx.WithTx(ctx, s.repo.DB(), func(tx pgx.Tx) error {
		var e error
		rows, e = s.repo.UpdateExpense(ctx, tx, args, sets)
		if e != nil {
			return e
		}
		if len(rows) == 0 {
			return apperror.NotFound("Data tidak ditemukan")
		}
		return s.repo.LogAudit(ctx, tx, auth.GymID, &auth.UserID, "expenses.updated", "expenses", nil, rows[0])
	})
	if err != nil {
		return nil, err
	}
	return rows[0], nil
}

func (s *Service) ChangeExpenseStatus(ctx context.Context, auth middleware.AuthUser, publicID, status, action string) (map[string]any, error) {
	id, err := parsePublicID(publicID)
	if err != nil {
		return nil, err
	}
	var rows []map[string]any
	err = txtx.WithTx(ctx, s.repo.DB(), func(tx pgx.Tx) error {
		var e error
		rows, e = s.repo.ChangeExpenseStatus(ctx, tx, auth.GymID, id, status)
		if e != nil {
			return e
		}
		if len(rows) == 0 {
			return apperror.NotFound("Data tidak ditemukan")
		}
		return s.repo.LogAudit(ctx, tx, auth.GymID, &auth.UserID, action, "expenses", nil, rows[0])
	})
	if err != nil {
		return nil, err
	}
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

func parsePublicID(raw string) (uuid.UUID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, apperror.BadRequest("public_id tidak valid")
	}
	return id, nil
}
