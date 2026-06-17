package workout

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
	"membership-gym/internal/modules/shared"
	"membership-gym/pkg/token"
	txtx "membership-gym/pkg/tx"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }
func (r *Repository) DB() *pgxpool.Pool          { return r.db }

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

func (r *Repository) Exec(ctx context.Context, query string, args ...any) error {
	_, err := r.db.Exec(ctx, query, args...)
	return err
}

func (r *Repository) ResolveID(ctx context.Context, table string, gymID int64, publicID uuid.UUID) (int64, error) {
	var id int64
	err := r.db.QueryRow(ctx, fmt.Sprintf(models.QueryResolveIDFormat, table), gymID, publicID).Scan(&id)
	return id, err
}

func (r *Repository) ListGeneric(ctx context.Context, cfg shared.ResourceConfig, where []string, args []any, limit, offset int) ([]map[string]any, int64, error) {
	total, err := r.Count(ctx, fmt.Sprintf(models.QueryGenericListCountFormat, cfg.Table, strings.Join(where, " AND ")), args...)
	if err != nil {
		return nil, 0, err
	}
	args = append(args, limit, offset)
	items, err := r.QueryMaps(ctx, fmt.Sprintf(models.QueryGenericListFormat, cfg.Select, cfg.Table, strings.Join(where, " AND "), len(args)-1, len(args)), args...)
	return items, total, err
}

func (r *Repository) DetailGeneric(ctx context.Context, cfg shared.ResourceConfig, gymID int64, publicID uuid.UUID) ([]map[string]any, error) {
	return r.QueryMaps(ctx, fmt.Sprintf(models.QueryGenericDetailFormat, cfg.Select, cfg.Table), gymID, publicID)
}

func (r *Repository) UpdateGeneric(ctx context.Context, db txtx.DBTX, cfg shared.ResourceConfig, args []any, sets []string) ([]map[string]any, error) {
	return r.QueryMapsWith(ctx, db, fmt.Sprintf(models.QueryGenericUpdate, cfg.Table, strings.Join(sets, ","), cfg.Select), args...)
}

func (r *Repository) ChangeStatusGeneric(ctx context.Context, db txtx.DBTX, cfg shared.ResourceConfig, gymID int64, publicID uuid.UUID, status any) ([]map[string]any, error) {
	return r.QueryMapsWith(ctx, db, fmt.Sprintf(models.QueryGenericChangeStatus, cfg.Table, cfg.StatusColumn, cfg.Select), gymID, publicID, status)
}

func (r *Repository) InsertMuscleGroup(ctx context.Context, db txtx.DBTX, args ...any) ([]map[string]any, error) {
	return r.QueryMapsWith(ctx, db, models.QueryInsertMuscleGroup, args...)
}

func (r *Repository) InsertWorkoutTemplate(ctx context.Context, db txtx.DBTX, args ...any) ([]map[string]any, error) {
	return r.QueryMapsWith(ctx, db, models.QueryInsertWorkoutTemplate, args...)
}

func (r *Repository) ActivePlanFeatures(ctx context.Context, gymID int64) ([]map[string]any, error) {
	return r.QueryMaps(ctx, models.QueryActivePlanFeatures, gymID)
}

func (r *Repository) LogAudit(ctx context.Context, db txtx.DBTX, gymID int64, userID *int64, action, entityType string, entityID *int64, payload any) error {
	data, _ := json.Marshal(payload)
	_, err := db.Exec(ctx, models.QueryAuditLogInsert, token.GeneratePublicID(), gymID, userID, action, entityType, entityID, data)
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
