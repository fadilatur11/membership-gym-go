package qrcode

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"membership-gym/config"
	"membership-gym/internal/middleware"
	apperror "membership-gym/pkg/errors"
	"membership-gym/pkg/token"
	txtx "membership-gym/pkg/tx"
)

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

func (s *Service) GetOrCreateQR(ctx context.Context, auth middleware.AuthUser, memberPublicID string, regenerate bool) (map[string]any, error) {
	memberPID, err := parsePublicID(memberPublicID)
	if err != nil {
		return nil, err
	}
	memberRows, err := s.repo.MemberForQR(ctx, auth.GymID, memberPID)
	if err != nil || len(memberRows) == 0 {
		return nil, apperror.NotFound("Member tidak ditemukan")
	}
	member := memberRows[0]
	if !regenerate {
		rows, err := s.repo.ActiveQR(ctx, auth.GymID, member["id"], member["public_id"], member["full_name"], s.cfg.PublicMemberURL)
		if err != nil {
			return nil, err
		}
		if len(rows) > 0 {
			return rows[0], nil
		}
	}
	var out []map[string]any
	err = txtx.WithTx(ctx, s.repo.DB(), func(tx pgx.Tx) error {
		if regenerate {
			if e := s.repo.RevokeActiveQRByMemberID(ctx, tx, auth.GymID, member["id"].(int64)); e != nil {
				return e
			}
		}
		var e error
		out, e = s.repo.InsertMemberQR(ctx, tx, token.GeneratePublicID(), auth.GymID, member["id"], token.GenerateQRCodeToken(), member["public_id"], member["full_name"], s.cfg.PublicMemberURL)
		if e != nil {
			return e
		}
		action := "qrcode.generated"
		if regenerate {
			action = "qrcode.regenerated"
		}
		return s.repo.LogAudit(ctx, tx, auth.GymID, &auth.UserID, action, "member_qrcodes", nil, out[0])
	})
	if err != nil {
		return nil, err
	}
	return out[0], nil
}

func (s *Service) RevokeQR(ctx context.Context, auth middleware.AuthUser, memberPublicID string) (map[string]any, error) {
	memberPID, err := parsePublicID(memberPublicID)
	if err != nil {
		return nil, err
	}
	var out []map[string]any
	err = txtx.WithTx(ctx, s.repo.DB(), func(tx pgx.Tx) error {
		var e error
		out, e = s.repo.RevokeQR(ctx, tx, auth.GymID, memberPID)
		if e != nil {
			return e
		}
		if len(out) == 0 {
			return apperror.NotFound("QR aktif tidak ditemukan")
		}
		return s.repo.LogAudit(ctx, tx, auth.GymID, &auth.UserID, "qrcode.revoked", "member_qrcodes", nil, out[0])
	})
	if err != nil {
		return nil, err
	}
	return out[0], nil
}

func parsePublicID(raw string) (uuid.UUID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, apperror.BadRequest("public_id tidak valid")
	}
	return id, nil
}
