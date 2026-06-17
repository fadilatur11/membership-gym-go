package audit

import (
	"context"
	"fmt"
	"strings"

	"membership-gym/internal/middleware"
	"membership-gym/pkg/pagination"
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

func (s *Service) ListAudit(ctx context.Context, auth middleware.AuthUser, p pagination.Params, q map[string]string) (PageResult, error) {
	where := []string{"gym_id=$1"}
	args := []any{auth.GymID}
	if entityType := strings.TrimSpace(q["entity_type"]); entityType != "" {
		args = append(args, entityType)
		where = append(where, fmt.Sprintf("entity_type=$%d", len(args)))
	}
	items, total, err := s.repo.ListAudit(ctx, where, args, p.Limit, p.Offset())
	if err != nil {
		return PageResult{}, err
	}
	return PageResult{Items: items, Meta: pagination.NewMeta(p.Page, p.Limit, total)}, nil
}
