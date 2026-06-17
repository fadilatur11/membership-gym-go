package gym

import (
	"context"
	"fmt"

	"membership-gym/internal/middleware"
	apperror "membership-gym/pkg/errors"
)

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

func (s *Service) GymProfile(ctx context.Context, auth middleware.AuthUser) (map[string]any, error) {
	rows, err := s.repo.GymProfile(ctx, auth.GymID)
	if err != nil || len(rows) == 0 {
		return nil, apperror.NotFound("Gym tidak ditemukan")
	}
	return rows[0], nil
}

func (s *Service) UpdateGym(ctx context.Context, auth middleware.AuthUser, req UpdateGymRequest) (map[string]any, error) {
	err := s.repo.UpdateGymProfile(ctx, auth.GymID, req.Name, req.Phone, req.Email, req.Address, req.Timezone, req.Currency)
	if err != nil {
		return nil, err
	}
	return s.GymProfile(ctx, auth)
}

func (s *Service) PublicGymContact(ctx context.Context, qrToken string) (map[string]any, error) {
	rows, err := s.repo.PublicGymContact(ctx, qrToken)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, apperror.NotFound("QR tidak valid")
	}
	return rows[0], nil
}
