package payment

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
	"membership-gym/pkg/invoice"
	"membership-gym/pkg/pagination"
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

func (s *Service) ListPayments(ctx context.Context, auth middleware.AuthUser, p pagination.Params, q map[string]string) (PageResult, error) {
	where := []string{"gym_id=$1"}
	args := []any{auth.GymID}
	if status := strings.TrimSpace(q["status"]); status != "" {
		args = append(args, status)
		where = append(where, fmt.Sprintf("status=$%d", len(args)))
	}
	if method := strings.TrimSpace(q["payment_method"]); method != "" {
		args = append(args, method)
		where = append(where, fmt.Sprintf("payment_method=$%d", len(args)))
	}
	items, total, err := s.repo.ListPayments(ctx, where, args, p.Limit, p.Offset())
	if err != nil {
		return PageResult{}, err
	}
	return PageResult{Items: items, Meta: pagination.NewMeta(p.Page, p.Limit, total)}, nil
}

func (s *Service) DetailPayment(ctx context.Context, auth middleware.AuthUser, publicID string) (map[string]any, error) {
	id, err := parsePublicID(publicID)
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.DetailPayment(ctx, auth.GymID, id)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, apperror.NotFound("Data tidak ditemukan")
	}
	return rows[0], nil
}

func (s *Service) CreatePayment(ctx context.Context, auth middleware.AuthUser, req CreatePaymentRequest) (map[string]any, error) {
	memberPID, err := parsePublicID(req.MemberPublicID)
	if err != nil {
		return nil, err
	}
	packagePID, err := parsePublicID(req.MembershipPackagePublicID)
	if err != nil {
		return nil, err
	}
	startDate, err := dateFromString(req.StartDate, datetime.TodayInTimezone(s.cfg.DefaultTimezone), "start_date")
	if err != nil {
		return nil, err
	}
	method, status := req.PaymentMethod, req.Status
	if method != "cash" && method != "transfer" && method != "qris" {
		return nil, apperror.BadRequest("payment_method tidak valid")
	}
	if status == "" {
		status = "paid"
	}
	if status != "paid" && status != "pending" {
		return nil, apperror.BadRequest("status pembayaran tidak valid")
	}
	rows, err := s.repo.MemberPackageForPayment(ctx, auth.GymID, memberPID, packagePID)
	if err != nil || len(rows) == 0 {
		return nil, apperror.BadRequest("Member atau package tidak valid")
	}
	base := rows[0]
	price, discount := int64Num(base["price"]), req.DiscountAmount
	if discount < 0 || discount > price {
		return nil, apperror.BadRequest("discount_amount tidak valid")
	}
	endDate := startDate.AddDate(0, 0, intNum(base["duration_days"])-1)
	now := datetime.NowInTimezone(s.cfg.DefaultTimezone)
	var out map[string]any
	err = txtx.WithTx(ctx, s.repo.DB(), func(tx pgx.Tx) error {
		subRows, e := s.repo.InsertPaymentSubscription(ctx, tx, token.GeneratePublicID(), auth.GymID, base["member_id"], base["package_id"], startDate, endDate)
		if e != nil {
			return e
		}
		seq, e := s.repo.PaymentSequence(ctx, tx, auth.GymID, now.Format("2006-01-02"))
		if e != nil {
			return e
		}
		invoiceNo := invoice.GenerateInvoiceNo(now, seq)
		payRows, e := s.repo.InsertPayment(ctx, tx, token.GeneratePublicID(), auth.GymID, base["member_id"], subRows[0]["id"], invoiceNo, method, price, discount, price-discount, status, now, req.Notes, auth.UserID)
		if e != nil {
			return e
		}
		out = payRows[0]
		out["member_name"] = base["member_name"]
		out["package_name"] = base["package_name"]
		out["subscription"] = map[string]any{"public_id": subRows[0]["public_id"], "start_date": subRows[0]["start_date"], "end_date": subRows[0]["end_date"], "status": subRows[0]["status"]}
		return s.repo.LogAudit(ctx, tx, auth.GymID, &auth.UserID, "payment.created", "payments", nil, out)
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Service) ChangePayment(ctx context.Context, auth middleware.AuthUser, publicID, status string) (map[string]any, error) {
	id, err := parsePublicID(publicID)
	if err != nil {
		return nil, err
	}
	action := "payment.cancelled"
	if status == "refunded" {
		action = "payment.refunded"
	}
	var out []map[string]any
	err = txtx.WithTx(ctx, s.repo.DB(), func(tx pgx.Tx) error {
		subID, e := s.repo.PaymentSubscriptionID(ctx, tx, auth.GymID, id)
		if e != nil {
			return e
		}
		outRows, e := s.repo.UpdatePaymentStatus(ctx, tx, auth.GymID, id, status)
		if e != nil {
			return e
		}
		if len(outRows) == 0 {
			return apperror.NotFound("Data tidak ditemukan")
		}
		out = outRows
		if subID != nil {
			if e = s.repo.CancelSubscriptionByID(ctx, tx, auth.GymID, *subID); e != nil {
				return e
			}
		}
		return s.repo.LogAudit(ctx, tx, auth.GymID, &auth.UserID, action, "payments", nil, out[0])
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
