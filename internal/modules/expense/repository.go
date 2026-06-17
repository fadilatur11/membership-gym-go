package expense

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

const expenseSelect = "public_id, expense_category_id, title, description, amount, expense_date, status, created_at, updated_at"

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

func (r *Repository) ListExpenses(ctx context.Context, where []string, args []any, limit, offset int) ([]map[string]any, int64, error) {
	total, err := r.Count(ctx, fmt.Sprintf(models.QueryGenericListCountFormat, "expenses", strings.Join(where, " AND ")), args...)
	if err != nil {
		return nil, 0, err
	}
	args = append(args, limit, offset)
	items, err := r.QueryMaps(ctx, fmt.Sprintf(models.QueryGenericListFormat, expenseSelect, "expenses", strings.Join(where, " AND "), len(args)-1, len(args)), args...)
	return items, total, err
}

func (r *Repository) DetailExpense(ctx context.Context, gymID int64, publicID uuid.UUID) ([]map[string]any, error) {
	return r.QueryMaps(ctx, fmt.Sprintf(models.QueryGenericDetailFormat, expenseSelect, "expenses"), gymID, publicID)
}

func (r *Repository) ResolveID(ctx context.Context, table string, gymID int64, publicID uuid.UUID) (int64, error) {
	var id int64
	err := r.db.QueryRow(ctx, fmt.Sprintf(models.QueryResolveIDFormat, table), gymID, publicID).Scan(&id)
	return id, err
}

func (r *Repository) InsertExpense(ctx context.Context, db txtx.DBTX, args ...any) ([]map[string]any, error) {
	return r.QueryMapsWith(ctx, db, models.QueryInsertExpense, args...)
}

func (r *Repository) UpdateExpense(ctx context.Context, db txtx.DBTX, args []any, sets []string) ([]map[string]any, error) {
	return r.QueryMapsWith(ctx, db, fmt.Sprintf(models.QueryGenericUpdate, "expenses", strings.Join(sets, ","), expenseSelect), args...)
}

func (r *Repository) ChangeExpenseStatus(ctx context.Context, db txtx.DBTX, gymID int64, publicID uuid.UUID, status string) ([]map[string]any, error) {
	return r.QueryMapsWith(ctx, db, fmt.Sprintf(models.QueryGenericChangeStatus, "expenses", "status", expenseSelect), gymID, publicID, status)
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
