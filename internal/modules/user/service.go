package user

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"membership-gym/internal/middleware"
	apperror "membership-gym/pkg/errors"
	"membership-gym/pkg/hash"
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

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
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

func (s *Service) ListUsers(ctx context.Context, auth middleware.AuthUser, p pagination.Params, q map[string]string) (PageResult, error) {
	where := []string{"gym_id=$1"}
	args := []any{auth.GymID}
	if search := strings.TrimSpace(q["search"]); search != "" {
		args = append(args, sqlutil.LikeSearch(search))
		where = append(where, fmt.Sprintf("(LOWER(name) LIKE $%d OR LOWER(email) LIKE $%d)", len(args), len(args)))
	}
	if role := strings.TrimSpace(q["role"]); role != "" {
		args = append(args, role)
		where = append(where, fmt.Sprintf("role=$%d", len(args)))
	}
	if active := strings.TrimSpace(q["is_active"]); active != "" {
		args = append(args, active)
		where = append(where, fmt.Sprintf("is_active=$%d", len(args)))
	}
	items, total, err := s.repo.ListUsers(ctx, where, args, p.Limit, p.Offset())
	if err != nil {
		return PageResult{}, err
	}
	return PageResult{Items: items, Meta: pagination.NewMeta(p.Page, p.Limit, total)}, nil
}

func (s *Service) DetailUser(ctx context.Context, auth middleware.AuthUser, publicID string) (map[string]any, error) {
	id, err := parsePublicID(publicID)
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.DetailUser(ctx, auth.GymID, id)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, apperror.NotFound("Data tidak ditemukan")
	}
	return rows[0], nil
}

func (s *Service) CreateUser(ctx context.Context, auth middleware.AuthUser, req CreateUserRequest) (map[string]any, error) {
	if !validRole(req.Role) {
		return nil, apperror.BadRequest("Role tidak valid")
	}
	if (auth.Role == "cashier" || auth.Role == "trainer") && (req.Role == "owner" || req.Role == "admin") {
		return nil, apperror.Forbidden("Role tidak diizinkan")
	}
	if len(req.Password) < 6 {
		return nil, apperror.BadRequest("Password minimal 6 karakter")
	}
	hashed, err := hash.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}
	var rows []map[string]any
	err = txtx.WithTx(ctx, s.repo.DB(), func(tx pgx.Tx) error {
		roleID, e := s.repo.RoleIDByCode(ctx, tx, auth.GymID, req.Role)
		if e != nil {
			return e
		}
		rows, e = s.repo.InsertUser(ctx, tx, token.GeneratePublicID(), auth.GymID, req.Name, strings.ToLower(req.Email), hashed, req.Role, roleID)
		if e != nil {
			return e
		}
		return s.repo.LogAudit(ctx, tx, auth.GymID, &auth.UserID, "user.created", "users", nil, rows[0])
	})
	if err != nil {
		return nil, err
	}
	return rows[0], nil
}

func (s *Service) UpdateUser(ctx context.Context, auth middleware.AuthUser, publicID string, req UpdateUserRequest) (map[string]any, error) {
	id, err := parsePublicID(publicID)
	if err != nil {
		return nil, err
	}
	sets, args := []string{}, []any{auth.GymID, id}
	if req.Name != "" {
		args = append(args, req.Name)
		sets = append(sets, fmt.Sprintf("name=$%d", len(args)))
	}
	if req.Email != "" {
		args = append(args, strings.ToLower(req.Email))
		sets = append(sets, fmt.Sprintf("email=$%d", len(args)))
	}
	if req.Role != "" {
		if !validRole(req.Role) {
			return nil, apperror.BadRequest("Role tidak valid")
		}
		args = append(args, req.Role)
		sets = append(sets, fmt.Sprintf("role=$%d", len(args)))
	}
	if req.IsActive != nil {
		args = append(args, *req.IsActive)
		sets = append(sets, fmt.Sprintf("is_active=$%d", len(args)))
	}
	if len(sets) == 0 {
		return s.DetailUser(ctx, auth, publicID)
	}
	sets = append(sets, "updated_at=NOW()")
	var rows []map[string]any
	err = txtx.WithTx(ctx, s.repo.DB(), func(tx pgx.Tx) error {
		var e error
		rows, e = s.repo.UpdateUser(ctx, tx, args, sets)
		if e != nil {
			return e
		}
		if len(rows) == 0 {
			return apperror.NotFound("Data tidak ditemukan")
		}
		return s.repo.LogAudit(ctx, tx, auth.GymID, &auth.UserID, "users.updated", "users", nil, rows[0])
	})
	if err != nil {
		return nil, err
	}
	return rows[0], nil
}

func (s *Service) DeactivateUser(ctx context.Context, auth middleware.AuthUser, publicID string) (map[string]any, error) {
	id, err := parsePublicID(publicID)
	if err != nil {
		return nil, err
	}
	var rows []map[string]any
	err = txtx.WithTx(ctx, s.repo.DB(), func(tx pgx.Tx) error {
		var e error
		rows, e = s.repo.ChangeUserStatus(ctx, tx, auth.GymID, id, false)
		if e != nil {
			return e
		}
		if len(rows) == 0 {
			return apperror.NotFound("Data tidak ditemukan")
		}
		return s.repo.LogAudit(ctx, tx, auth.GymID, &auth.UserID, "user.deactivated", "users", nil, rows[0])
	})
	if err != nil {
		return nil, err
	}
	return rows[0], nil
}

func validRole(role string) bool {
	return role == "owner" || role == "admin" || role == "cashier" || role == "trainer"
}

func parsePublicID(raw string) (uuid.UUID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, apperror.BadRequest("public_id tidak valid")
	}
	return id, nil
}
