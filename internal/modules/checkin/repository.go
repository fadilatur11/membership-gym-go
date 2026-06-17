package checkin

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
)

const checkinSelect = "public_id, member_id, subscription_id, checkin_date, checkin_at, source, status, notes, created_at, updated_at"

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) QueryMaps(ctx context.Context, query string, args ...any) ([]map[string]any, error) {
	rows, err := r.db.Query(ctx, query, args...)
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

func (r *Repository) ListCheckins(ctx context.Context, where []string, args []any, limit, offset int) ([]map[string]any, int64, error) {
	countQuery := fmt.Sprintf(models.QueryGenericListCountFormat, "member_checkins", strings.Join(where, " AND "))
	total, err := r.Count(ctx, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	args = append(args, limit, offset)
	query := fmt.Sprintf(models.QueryGenericListFormat, checkinSelect, "member_checkins", strings.Join(where, " AND "), len(args)-1, len(args))
	items, err := r.QueryMaps(ctx, query, args...)
	return items, total, err
}

func (r *Repository) CheckinQR(ctx context.Context, scoped bool, args ...any) ([]map[string]any, error) {
	scopeSQL := ""
	if scoped {
		scopeSQL = " AND q.gym_id=$2"
	}
	return r.QueryMaps(ctx, fmt.Sprintf(models.QueryCheckinQRBase, scopeSQL), args...)
}

func (r *Repository) ActiveMemberSubscription(ctx context.Context, gymID, memberID int64, today time.Time) ([]map[string]any, error) {
	return r.QueryMaps(ctx, models.QueryActiveMemberSubscription, gymID, memberID, today)
}

func (r *Repository) ValidCheckinToday(ctx context.Context, gymID, memberID int64, today time.Time) ([]map[string]any, error) {
	return r.QueryMaps(ctx, models.QueryValidCheckinToday, gymID, memberID, today)
}

func (r *Repository) ManualCheckinMember(ctx context.Context, gymID int64, memberPublicID uuid.UUID) ([]map[string]any, error) {
	return r.QueryMaps(ctx, models.QueryManualCheckinMember, gymID, memberPublicID)
}

func (r *Repository) ManualActiveSubscription(ctx context.Context, gymID int64, memberID any, today time.Time) ([]map[string]any, error) {
	return r.QueryMaps(ctx, models.QueryManualActiveSubscription, gymID, memberID, today)
}

func (r *Repository) InsertMemberCheckin(ctx context.Context, args ...any) error {
	_, err := r.db.Exec(ctx, models.QueryInsertMemberCheckin, args...)
	return err
}

func (r *Repository) QRCodeMemberIDs(ctx context.Context, qrToken string) ([]map[string]any, error) {
	return r.QueryMaps(ctx, models.QueryQRCodeMemberIDs, qrToken)
}

func (r *Repository) CountMemberCheckins(ctx context.Context, gymID, memberID any) (int64, error) {
	return r.Count(ctx, models.QueryCountMemberCheckins, gymID, memberID)
}

func (r *Repository) PublicCheckinHistory(ctx context.Context, gymID, memberID any, limit, offset int) ([]map[string]any, error) {
	return r.QueryMaps(ctx, models.QueryPublicCheckinHistory, gymID, memberID, limit, offset)
}

func (r *Repository) ActivePlanFeatures(ctx context.Context, gymID int64) ([]map[string]any, error) {
	return r.QueryMaps(ctx, models.QueryActivePlanFeatures, gymID)
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
