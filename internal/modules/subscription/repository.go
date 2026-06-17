package subscription

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"membership-gym/internal/models"
	txtx "membership-gym/pkg/tx"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) DB() *pgxpool.Pool { return r.db }

func (r *Repository) QueryMaps(ctx context.Context, query string, args ...any) ([]map[string]any, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return scanMaps(rows)
}

func (r *Repository) QueryMapsWith(ctx context.Context, db txtx.DBTX, query string, args ...any) ([]map[string]any, error) {
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return scanMaps(rows)
}

func (r *Repository) Count(ctx context.Context, query string, args ...any) (int64, error) {
	var total int64
	err := r.db.QueryRow(ctx, query, args...).Scan(&total)
	return total, err
}

func (r *Repository) ListSubscriptions(ctx context.Context, where []string, args []any, limit, offset int) ([]map[string]any, int64, error) {
	countQuery := fmt.Sprintf(models.QueryGenericListCountFormat, "subscriptions", strings.Join(where, " AND "))
	total, err := r.Count(ctx, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	args = append(args, limit, offset)
	query := fmt.Sprintf(models.QueryGenericListFormat, subscriptionSelect, "subscriptions", strings.Join(where, " AND "), len(args)-1, len(args))
	items, err := r.QueryMaps(ctx, query, args...)
	return items, total, err
}

func (r *Repository) DetailSubscription(ctx context.Context, gymID int64, publicID uuid.UUID) ([]map[string]any, error) {
	return r.QueryMaps(ctx, fmt.Sprintf(models.QueryGenericDetailFormat, subscriptionSelect, "subscriptions"), gymID, publicID)
}

func (r *Repository) ChangeSubscriptionStatus(ctx context.Context, db txtx.DBTX, gymID int64, publicID uuid.UUID, status string) ([]map[string]any, error) {
	query := fmt.Sprintf(models.QueryGenericChangeStatus, "subscriptions", "status", subscriptionSelect)
	return r.QueryMapsWith(ctx, db, query, gymID, publicID, status)
}

func (r *Repository) ListSaasPlans(ctx context.Context) ([]map[string]any, error) {
	return r.QueryMaps(ctx, models.QueryListSaasPlans)
}

func (r *Repository) CurrentGymSubscription(ctx context.Context, gymID int64) ([]map[string]any, error) {
	return r.QueryMaps(ctx, models.QueryCurrentGymSubscription, gymID)
}

func (r *Repository) SaasPlanByPublicID(ctx context.Context, publicID uuid.UUID) ([]map[string]any, error) {
	return r.QueryMaps(ctx, models.QuerySaasPlanByPublicID, publicID)
}

func (r *Repository) CancelCurrentGymSubscriptions(ctx context.Context, db txtx.DBTX, gymID int64) error {
	_, err := db.Exec(ctx, models.QueryCancelCurrentGymSubscriptions, gymID)
	return err
}

func (r *Repository) InsertOwnerGymSubscription(ctx context.Context, db txtx.DBTX, args ...any) ([]map[string]any, error) {
	return r.QueryMapsWith(ctx, db, models.QueryInsertOwnerGymSubscription, args...)
}

func (r *Repository) GymSubscriptionPaymentSequence(ctx context.Context, db txtx.DBTX, gymID int64, date string) (int, error) {
	var seq int
	err := db.QueryRow(ctx, models.QueryGymSubscriptionPaymentSequence, gymID, date).Scan(&seq)
	return seq, err
}

func (r *Repository) InsertGymSubscriptionPayment(ctx context.Context, db txtx.DBTX, args ...any) ([]map[string]any, error) {
	return r.QueryMapsWith(ctx, db, models.QueryInsertGymSubscriptionPayment, args...)
}

func (r *Repository) ActivePlanFeatures(ctx context.Context, gymID int64) ([]map[string]any, error) {
	return r.QueryMaps(ctx, models.QueryActivePlanFeatures, gymID)
}

func (r *Repository) MemberPackageForSubscription(ctx context.Context, gymID int64, memberPID, packagePID uuid.UUID) ([]map[string]any, error) {
	return r.QueryMaps(ctx, models.QueryMemberPackageForSubscription, gymID, memberPID, packagePID)
}

func (r *Repository) InsertSubscription(ctx context.Context, db txtx.DBTX, args ...any) ([]map[string]any, error) {
	return r.QueryMapsWith(ctx, db, models.QueryInsertSubscription, args...)
}

func (r *Repository) LogAudit(ctx context.Context, db txtx.DBTX, gymID int64, userID *int64, action, entityType string, entityID *int64, payload any) error {
	data, _ := json.Marshal(payload)
	_, err := db.Exec(ctx, models.QueryAuditLogInsert, uuid.New(), gymID, userID, action, entityType, entityID, data)
	return err
}

func scanMaps(rows pgx.Rows) ([]map[string]any, error) {
	defer rows.Close()
	fields := rows.FieldDescriptions()
	result := []map[string]any{}
	for rows.Next() {
		values := make([]any, len(fields))
		ptrs := make([]any, len(fields))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		item := map[string]any{}
		for i, field := range fields {
			key := string(field.Name)
			switch value := values[i].(type) {
			case time.Time:
				if value.Hour() == 0 && value.Minute() == 0 && value.Second() == 0 {
					item[key] = value.Format("2006-01-02")
				} else {
					item[key] = value.Format("2006-01-02 15:04:05")
				}
			case []byte:
				var jsonValue any
				if json.Valid(value) && json.Unmarshal(value, &jsonValue) == nil {
					item[key] = jsonValue
				} else {
					item[key] = string(value)
				}
			case [16]byte:
				item[key] = uuid.UUID(value)
			default:
				item[key] = value
			}
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
