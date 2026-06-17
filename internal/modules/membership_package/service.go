package membership_package

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

func (s *Service) ListPackages(ctx context.Context, auth middleware.AuthUser, p pagination.Params, q map[string]string) (PageResult, error) {
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
	items, total, err := s.repo.ListPackages(ctx, where, args, p.Limit, p.Offset())
	if err != nil {
		return PageResult{}, err
	}
	return PageResult{Items: items, Meta: pagination.NewMeta(p.Page, p.Limit, total)}, nil
}

func (s *Service) DetailPackage(ctx context.Context, auth middleware.AuthUser, publicID string) (map[string]any, error) {
	id, err := parsePublicID(publicID)
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.DetailPackage(ctx, auth.GymID, id)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, apperror.NotFound("Data tidak ditemukan")
	}
	return rows[0], nil
}

func (s *Service) CreatePackage(ctx context.Context, auth middleware.AuthUser, req CreateMembershipPackageRequest) (map[string]any, error) {
	if req.DurationDays <= 0 || req.Price < 0 {
		return nil, apperror.BadRequest("duration_days harus > 0 dan price >= 0")
	}
	var rows []map[string]any
	err := txtx.WithTx(ctx, s.repo.DB(), func(tx pgx.Tx) error {
		var e error
		rows, e = s.repo.InsertMembershipPackage(ctx, tx, token.GeneratePublicID(), auth.GymID, req.Name, req.DurationDays, req.Price, req.Description)
		if e != nil {
			return e
		}
		return s.repo.LogAudit(ctx, tx, auth.GymID, &auth.UserID, "package.created", "membership_packages", nil, rows[0])
	})
	if err != nil {
		return nil, err
	}
	return rows[0], nil
}

func (s *Service) UpdatePackage(ctx context.Context, auth middleware.AuthUser, publicID string, req UpdateMembershipPackageRequest) (map[string]any, error) {
	id, err := parsePublicID(publicID)
	if err != nil {
		return nil, err
	}
	sets, args := []string{}, []any{auth.GymID, id}
	if req.Name != "" {
		args = append(args, req.Name)
		sets = append(sets, fmt.Sprintf("name=$%d", len(args)))
	}
	if req.DurationDays > 0 {
		args = append(args, req.DurationDays)
		sets = append(sets, fmt.Sprintf("duration_days=$%d", len(args)))
	}
	if req.Price > 0 {
		args = append(args, req.Price)
		sets = append(sets, fmt.Sprintf("price=$%d", len(args)))
	}
	if req.Description != "" {
		args = append(args, req.Description)
		sets = append(sets, fmt.Sprintf("description=$%d", len(args)))
	}
	if req.IsActive != nil {
		args = append(args, *req.IsActive)
		sets = append(sets, fmt.Sprintf("is_active=$%d", len(args)))
	}
	if len(sets) == 0 {
		return s.DetailPackage(ctx, auth, publicID)
	}
	sets = append(sets, "updated_at=NOW()")
	var rows []map[string]any
	err = txtx.WithTx(ctx, s.repo.DB(), func(tx pgx.Tx) error {
		var e error
		rows, e = s.repo.UpdatePackage(ctx, tx, args, sets)
		if e != nil {
			return e
		}
		if len(rows) == 0 {
			return apperror.NotFound("Data tidak ditemukan")
		}
		return s.repo.LogAudit(ctx, tx, auth.GymID, &auth.UserID, "membership_packages.updated", "membership_packages", nil, rows[0])
	})
	if err != nil {
		return nil, err
	}
	return rows[0], nil
}

func (s *Service) DeactivatePackage(ctx context.Context, auth middleware.AuthUser, publicID string) (map[string]any, error) {
	id, err := parsePublicID(publicID)
	if err != nil {
		return nil, err
	}
	var rows []map[string]any
	err = txtx.WithTx(ctx, s.repo.DB(), func(tx pgx.Tx) error {
		var e error
		rows, e = s.repo.ChangePackageStatus(ctx, tx, auth.GymID, id, false)
		if e != nil {
			return e
		}
		if len(rows) == 0 {
			return apperror.NotFound("Data tidak ditemukan")
		}
		return s.repo.LogAudit(ctx, tx, auth.GymID, &auth.UserID, "package.deactivated", "membership_packages", nil, rows[0])
	})
	if err != nil {
		return nil, err
	}
	return rows[0], nil
}

func (s *Service) PublicPackages(ctx context.Context, qrToken string) ([]map[string]any, error) {
	return s.repo.PublicPackages(ctx, qrToken)
}

func parsePublicID(raw string) (uuid.UUID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, apperror.BadRequest("public_id tidak valid")
	}
	return id, nil
}
