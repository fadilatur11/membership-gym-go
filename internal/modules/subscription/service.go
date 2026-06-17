package subscription

import (
	"context"
	"fmt"
	"strconv"
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

const subscriptionSelect = "public_id, member_id, membership_package_id, start_date, end_date, status, source, created_at, updated_at"

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

func (s *Service) ListSaasPlans(ctx context.Context) ([]map[string]any, error) {
	return s.repo.ListSaasPlans(ctx)
}

func (s *Service) CurrentGymSubscription(ctx context.Context, auth middleware.AuthUser) (map[string]any, error) {
	rows, err := s.repo.CurrentGymSubscription(ctx, auth.GymID)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, apperror.NotFound("Subscription gym belum ada")
	}
	return rows[0], nil
}

func (s *Service) CreateGymSubscription(ctx context.Context, auth middleware.AuthUser, req CreateGymSubscriptionRequest) (map[string]any, error) {
	planPID, err := parsePublicID(req.SaasPlanPublicID)
	if err != nil {
		return nil, err
	}
	startDate, err := dateFromString(req.StartDate, datetime.TodayInTimezone(s.cfg.DefaultTimezone))
	if err != nil {
		return nil, err
	}
	planRows, err := s.repo.SaasPlanByPublicID(ctx, planPID)
	if err != nil {
		return nil, err
	}
	if len(planRows) == 0 {
		return nil, apperror.BadRequest("Subscription plan tidak valid")
	}
	plan := planRows[0]
	price := int64Num(plan["price"])
	method := req.PaymentMethod
	status := req.Status
	if status == "" {
		status = "paid"
	}
	if price == 0 {
		method = "free"
		status = "paid"
	}
	if method == "" {
		method = "transfer"
	}
	if method != "free" && method != "cash" && method != "transfer" && method != "qris" {
		return nil, apperror.BadRequest("payment_method tidak valid")
	}
	if status != "paid" && status != "pending" {
		return nil, apperror.BadRequest("status pembayaran tidak valid")
	}
	if method == "free" && price > 0 {
		return nil, apperror.BadRequest("payment_method free hanya untuk plan gratis")
	}
	endDate := startDate.AddDate(0, 0, intNum(plan["duration_days"])-1)
	subStatus := "active"
	if status == "pending" {
		subStatus = "pending"
	}
	if fmt.Sprint(plan["code"]) == "basic" && price == 0 {
		subStatus = "trialing"
	}
	autoRenew := false
	if req.AutoRenew != nil {
		autoRenew = *req.AutoRenew
	}
	now := datetime.NowInTimezone(s.cfg.DefaultTimezone)
	var out map[string]any
	err = txtx.WithTx(ctx, s.repo.DB(), func(tx pgx.Tx) error {
		if e := s.repo.CancelCurrentGymSubscriptions(ctx, tx, auth.GymID); e != nil {
			return e
		}
		subRows, e := s.repo.InsertOwnerGymSubscription(ctx, tx, token.GeneratePublicID(), auth.GymID, plan["id"], startDate, endDate, subStatus, autoRenew, auth.UserID)
		if e != nil {
			return e
		}
		seq, e := s.repo.GymSubscriptionPaymentSequence(ctx, tx, auth.GymID, now.Format("2006-01-02"))
		if e != nil {
			return e
		}
		invoiceNo := fmt.Sprintf("SAAS-%s-%04d", now.Format("20060102"), seq)
		var paidAt any
		if status == "paid" {
			paidAt = now
		}
		payRows, e := s.repo.InsertGymSubscriptionPayment(ctx, tx, token.GeneratePublicID(), auth.GymID, subRows[0]["id"], invoiceNo, method, price, plan["currency"], status, paidAt, req.Notes, auth.UserID)
		if e != nil {
			return e
		}
		out = subRows[0]
		out["plan"] = map[string]any{
			"public_id": plan["public_id"], "code": plan["code"], "name": plan["name"],
			"duration_days": plan["duration_days"], "price": plan["price"], "currency": plan["currency"],
			"billing_cycle": plan["billing_cycle"], "features": plan["features"], "limits": plan["limits"],
		}
		out["payment"] = payRows[0]
		return s.repo.LogAudit(ctx, tx, auth.GymID, &auth.UserID, "gym_subscription.created", "gym_subscriptions", nil, out)
	})
	if err != nil {
		return nil, err
	}
	return out, nil
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

func (s *Service) ListSubscriptions(ctx context.Context, auth middleware.AuthUser, p pagination.Params, q map[string]string) (PageResult, error) {
	where := []string{"gym_id=$1"}
	args := []any{auth.GymID}
	if status := strings.TrimSpace(q["status"]); status != "" {
		args = append(args, status)
		where = append(where, fmt.Sprintf("status=$%d", len(args)))
	}
	if search := strings.TrimSpace(q["search"]); search != "" {
		args = append(args, sqlutil.LikeSearch(search))
		where = append(where, fmt.Sprintf("(source LIKE $%d OR status LIKE $%d)", len(args), len(args)))
	}
	items, total, err := s.repo.ListSubscriptions(ctx, where, args, p.Limit, p.Offset())
	if err != nil {
		return PageResult{}, err
	}
	return PageResult{Items: items, Meta: pagination.NewMeta(p.Page, p.Limit, total)}, nil
}

func (s *Service) DetailSubscription(ctx context.Context, auth middleware.AuthUser, publicID string) (map[string]any, error) {
	id, err := parsePublicID(publicID)
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.DetailSubscription(ctx, auth.GymID, id)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, apperror.NotFound("Data tidak ditemukan")
	}
	return rows[0], nil
}

func (s *Service) CreateSubscription(ctx context.Context, auth middleware.AuthUser, req CreateSubscriptionRequest, source string) (map[string]any, error) {
	memberPID, err := parsePublicID(req.MemberPublicID)
	if err != nil {
		return nil, err
	}
	packagePID, err := parsePublicID(req.MembershipPackagePublicID)
	if err != nil {
		return nil, err
	}
	startDate, err := dateFromString(req.StartDate, datetime.TodayInTimezone(s.cfg.DefaultTimezone))
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.MemberPackageForSubscription(ctx, auth.GymID, memberPID, packagePID)
	if err != nil || len(rows) == 0 {
		return nil, apperror.BadRequest("Member atau package tidak valid")
	}
	row := rows[0]
	endDate := startDate.AddDate(0, 0, intNum(row["duration_days"])-1)
	var out []map[string]any
	err = txtx.WithTx(ctx, s.repo.DB(), func(tx pgx.Tx) error {
		var e error
		out, e = s.repo.InsertSubscription(ctx, tx, token.GeneratePublicID(), auth.GymID, row["member_id"], row["package_id"], startDate, endDate, source)
		if e != nil {
			return e
		}
		out[0]["member_name"] = row["member_name"]
		out[0]["package_name"] = row["package_name"]
		return s.repo.LogAudit(ctx, tx, auth.GymID, &auth.UserID, "subscription.created", "subscriptions", nil, out[0])
	})
	if err != nil {
		return nil, err
	}
	return out[0], nil
}

func (s *Service) ChangeSubscriptionStatus(ctx context.Context, auth middleware.AuthUser, publicID string, status string, action string) (map[string]any, error) {
	id, err := parsePublicID(publicID)
	if err != nil {
		return nil, err
	}
	var rows []map[string]any
	err = txtx.WithTx(ctx, s.repo.DB(), func(tx pgx.Tx) error {
		var e error
		rows, e = s.repo.ChangeSubscriptionStatus(ctx, tx, auth.GymID, id, status)
		if e != nil {
			return e
		}
		if len(rows) == 0 {
			return apperror.NotFound("Data tidak ditemukan")
		}
		return s.repo.LogAudit(ctx, tx, auth.GymID, &auth.UserID, action, "subscriptions", nil, rows[0])
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

func dateFromString(raw string, fallback time.Time) (time.Time, error) {
	if strings.TrimSpace(raw) == "" {
		return fallback, nil
	}
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return time.Time{}, apperror.BadRequest("Format tanggal harus YYYY-MM-DD")
	}
	return t, nil
}

func intNum(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case int32:
		return int(n)
	case float64:
		return int(n)
	default:
		i, _ := strconv.Atoi(fmt.Sprint(v))
		return i
	}
}

func int64Num(v any) int64 {
	switch n := v.(type) {
	case int64:
		return n
	case int:
		return int64(n)
	case int32:
		return int64(n)
	case float64:
		return int64(n)
	default:
		i, _ := strconv.ParseInt(fmt.Sprint(v), 10, 64)
		return i
	}
}
