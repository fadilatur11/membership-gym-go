package gym

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"membership-gym/internal/models"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

func (r *Repository) QueryMaps(ctx context.Context, query string, args ...any) ([]map[string]any, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return scanMaps(rows)
}

func (r *Repository) GymProfile(ctx context.Context, gymID int64) ([]map[string]any, error) {
	return r.QueryMaps(ctx, models.QueryGymProfile, gymID)
}

func (r *Repository) UpdateGymProfile(ctx context.Context, args ...any) error {
	_, err := r.QueryMaps(ctx, models.QueryUpdateGymProfile, args...)
	return err
}

func (r *Repository) PublicGymContact(ctx context.Context, qrToken string) ([]map[string]any, error) {
	return r.QueryMaps(ctx, models.QueryPublicGymContact, qrToken)
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
