package user

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
	"membership-gym/pkg/token"
	txtx "membership-gym/pkg/tx"
)

const userSelect = "public_id, name, email, role, is_active, last_login_at, created_at, updated_at"

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

func (r *Repository) ListUsers(ctx context.Context, where []string, args []any, limit, offset int) ([]map[string]any, int64, error) {
	countQuery := fmt.Sprintf(models.QueryGenericListCountFormat, "users", strings.Join(where, " AND "))
	total, err := r.Count(ctx, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	args = append(args, limit, offset)
	query := fmt.Sprintf(models.QueryGenericListFormat, userSelect, "users", strings.Join(where, " AND "), len(args)-1, len(args))
	items, err := r.QueryMaps(ctx, query, args...)
	return items, total, err
}

func (r *Repository) DetailUser(ctx context.Context, gymID int64, publicID uuid.UUID) ([]map[string]any, error) {
	return r.QueryMaps(ctx, fmt.Sprintf(models.QueryGenericDetailFormat, userSelect, "users"), gymID, publicID)
}

func (r *Repository) UpdateUser(ctx context.Context, db txtx.DBTX, args []any, sets []string) ([]map[string]any, error) {
	query := fmt.Sprintf(models.QueryGenericUpdate, "users", strings.Join(sets, ","), userSelect)
	return r.QueryMapsWith(ctx, db, query, args...)
}

func (r *Repository) ChangeUserStatus(ctx context.Context, db txtx.DBTX, gymID int64, publicID uuid.UUID, active bool) ([]map[string]any, error) {
	query := fmt.Sprintf(models.QueryGenericChangeStatus, "users", "is_active", userSelect)
	return r.QueryMapsWith(ctx, db, query, gymID, publicID, active)
}

func (r *Repository) RoleIDByCode(ctx context.Context, db txtx.DBTX, gymID int64, code string) (int64, error) {
	var id int64
	err := db.QueryRow(ctx, models.QueryRoleIDByCode, gymID, code).Scan(&id)
	return id, err
}

func (r *Repository) InsertUser(ctx context.Context, db txtx.DBTX, args ...any) ([]map[string]any, error) {
	return r.QueryMapsWith(ctx, db, models.QueryInsertUser, args...)
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
