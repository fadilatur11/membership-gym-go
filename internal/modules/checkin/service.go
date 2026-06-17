package checkin

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"membership-gym/config"
	"membership-gym/internal/middleware"
	"membership-gym/pkg/datetime"
	apperror "membership-gym/pkg/errors"
	"membership-gym/pkg/pagination"
	"membership-gym/pkg/token"
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

func (s *Service) CheckinByQR(ctx context.Context, auth *middleware.AuthUser, qrToken string) (map[string]any, error) {
	if qrToken == "" {
		return nil, apperror.BadRequest("qr_token wajib diisi")
	}
	scoped, args := false, []any{qrToken}
	if auth != nil {
		scoped = true
		args = append(args, auth.GymID)
	}
	rows, err := s.repo.CheckinQR(ctx, scoped, args...)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, apperror.NotFound("QR tidak valid")
	}
	base := rows[0]
	gymID, memberID := base["gym_id"].(int64), base["member_id"].(int64)
	today := datetime.TodayInTimezone(fmt.Sprint(base["timezone"]))
	now := datetime.NowInTimezone(fmt.Sprint(base["timezone"]))
	if fmt.Sprint(base["member_status"]) != "active" {
		_ = s.recordCheckin(ctx, gymID, memberID, nil, today, now, "qr", "inactive_member", nil, "")
		return nil, apperror.BadRequest("Member tidak aktif")
	}
	subRows, err := s.repo.ActiveMemberSubscription(ctx, gymID, memberID, today)
	if err != nil {
		return nil, err
	}
	if len(subRows) == 0 {
		_ = s.recordCheckin(ctx, gymID, memberID, nil, today, now, "qr", "expired", nil, "")
		base["status"] = "expired"
		return nil, apperror.BadRequest("Membership sudah expired")
	}
	dup, err := s.repo.ValidCheckinToday(ctx, gymID, memberID, today)
	if err != nil {
		return nil, err
	}
	if len(dup) > 0 {
		_ = s.recordCheckin(ctx, gymID, memberID, nil, today, now, "qr", "duplicate", nil, "")
		return nil, apperror.Conflict("Member sudah check-in hari ini")
	}
	var scannedBy *int64
	if auth != nil {
		scannedBy = &auth.UserID
	}
	subID := subRows[0]["id"].(int64)
	if err := s.recordCheckin(ctx, gymID, memberID, &subID, today, now, "qr", "valid", scannedBy, ""); err != nil {
		return nil, err
	}
	base["checkin_at"] = now.Format("2006-01-02 15:04:05")
	base["subscription_end_date"] = subRows[0]["end_date"]
	base["status"] = "valid"
	delete(base, "gym_id")
	delete(base, "member_id")
	delete(base, "timezone")
	return base, nil
}

func (s *Service) ManualCheckin(ctx context.Context, auth middleware.AuthUser, memberPublicID string, notes string) (map[string]any, error) {
	memberPID, err := parsePublicID(memberPublicID)
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.ManualCheckinMember(ctx, auth.GymID, memberPID)
	if err != nil || len(rows) == 0 {
		return nil, apperror.NotFound("Member tidak ditemukan")
	}
	base := rows[0]
	if fmt.Sprint(base["member_status"]) != "active" {
		return nil, apperror.BadRequest("Member tidak aktif")
	}
	today := datetime.TodayInTimezone(s.cfg.DefaultTimezone)
	subRows, err := s.repo.ManualActiveSubscription(ctx, auth.GymID, base["id"], today)
	if err != nil || len(subRows) == 0 {
		return nil, apperror.BadRequest("Membership sudah expired")
	}
	dup, _ := s.repo.ValidCheckinToday(ctx, auth.GymID, base["id"].(int64), today)
	if len(dup) > 0 {
		return nil, apperror.Conflict("Member sudah check-in hari ini")
	}
	subID := subRows[0]["id"].(int64)
	now := datetime.NowInTimezone(s.cfg.DefaultTimezone)
	if err := s.recordCheckin(ctx, auth.GymID, base["id"].(int64), &subID, today, now, "manual", "valid", &auth.UserID, notes); err != nil {
		return nil, err
	}
	base["checkin_at"] = now.Format("2006-01-02 15:04:05")
	base["subscription_end_date"] = subRows[0]["end_date"]
	base["status"] = "valid"
	delete(base, "id")
	return base, nil
}

func (s *Service) ListCheckins(ctx context.Context, auth middleware.AuthUser, p pagination.Params, q map[string]string) (PageResult, error) {
	where := []string{"gym_id=$1"}
	args := []any{auth.GymID}
	for key, column := range map[string]string{"status": "status", "source": "source"} {
		if value := strings.TrimSpace(q[key]); value != "" {
			args = append(args, value)
			where = append(where, fmt.Sprintf("%s=$%d", column, len(args)))
		}
	}
	items, total, err := s.repo.ListCheckins(ctx, where, args, p.Limit, p.Offset())
	if err != nil {
		return PageResult{}, err
	}
	return PageResult{Items: items, Meta: pagination.NewMeta(p.Page, p.Limit, total)}, nil
}

func (s *Service) PublicCheckins(ctx context.Context, qrToken string, p pagination.Params) (PageResult, error) {
	base, err := s.repo.QRCodeMemberIDs(ctx, qrToken)
	if err != nil || len(base) == 0 {
		return PageResult{}, apperror.NotFound("QR tidak valid")
	}
	gymID, memberID := base[0]["gym_id"], base[0]["member_id"]
	total, err := s.repo.CountMemberCheckins(ctx, gymID, memberID)
	if err != nil {
		return PageResult{}, err
	}
	items, err := s.repo.PublicCheckinHistory(ctx, gymID, memberID, p.Limit, p.Offset())
	if err != nil {
		return PageResult{}, err
	}
	return PageResult{Items: items, Meta: pagination.NewMeta(p.Page, p.Limit, total)}, nil
}

func (s *Service) recordCheckin(ctx context.Context, gymID, memberID int64, subID *int64, date, at time.Time, source, status string, scannedBy *int64, notes string) error {
	return s.repo.InsertMemberCheckin(ctx, token.GeneratePublicID(), gymID, memberID, subID, date, at, source, status, scannedBy, notes)
}

func parsePublicID(raw string) (uuid.UUID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, apperror.BadRequest("public_id tidak valid")
	}
	return id, nil
}
