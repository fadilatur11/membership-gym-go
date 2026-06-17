package auth

import (
	"context"
	"database/sql"
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

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) DB() *pgxpool.Pool { return r.db }

func (r *Repository) FindLoginUser(ctx context.Context, where string, args ...any) ([]map[string]any, error) {
	return r.queryMaps(ctx, fmt.Sprintf(models.QueryLoginUserBase, where), args...)
}

func (r *Repository) TouchLastLogin(ctx context.Context, userID int64) error {
	_, err := r.db.Exec(ctx, models.QueryTouchLastLogin, userID)
	return err
}

func (r *Repository) FindRefreshUser(ctx context.Context, refreshValue string) ([]map[string]any, error) {
	return r.queryMaps(ctx, models.QueryRefreshUser, refreshValue)
}

func (r *Repository) FindIssueTokenUser(ctx context.Context, userID, gymID int64) ([]map[string]any, error) {
	return r.queryMaps(ctx, models.QueryIssueTokenUser, userID, gymID)
}

func (r *Repository) CountUsersByEmail(ctx context.Context, email string) (int64, error) {
	var total int64
	err := r.db.QueryRow(ctx, models.QueryCountUsersByEmail, email).Scan(&total)
	return total, err
}

func (r *Repository) FindGoogleUser(ctx context.Context, googleSub, email string) ([]map[string]any, error) {
	return r.queryMaps(ctx, models.QueryGoogleUserBySubOrEmail, googleSub, email)
}

func (r *Repository) LinkGoogleUser(ctx context.Context, userID int64, googleSub, avatarURL string) error {
	_, err := r.db.Exec(ctx, models.QueryLinkGoogleUser, userID, googleSub, avatarURL)
	return err
}

func (r *Repository) InsertGymReturningID(ctx context.Context, db txtx.DBTX, args ...any) error {
	return db.QueryRow(ctx, models.QueryInsertGymReturningID, args[:len(args)-1]...).Scan(args[len(args)-1])
}

func (r *Repository) InsertOwnerUserReturningID(ctx context.Context, db txtx.DBTX, args ...any) error {
	return db.QueryRow(ctx, models.QueryInsertOwnerUserReturningID, args[:len(args)-1]...).Scan(args[len(args)-1])
}

func (r *Repository) InsertDefaultRole(ctx context.Context, db txtx.DBTX, args ...any) error {
	_, err := db.Exec(ctx, models.QueryInsertDefaultRole, args...)
	return err
}

func (r *Repository) OwnerRoleID(ctx context.Context, db txtx.DBTX, gymID int64) (int64, error) {
	var id int64
	err := db.QueryRow(ctx, models.QueryOwnerRoleID, gymID).Scan(&id)
	return id, err
}

func (r *Repository) BasicSaasPlan(ctx context.Context, db txtx.DBTX) (int64, int, error) {
	var planID int64
	var duration int
	err := db.QueryRow(ctx, models.QueryBasicSaasPlan).Scan(&planID, &duration)
	return planID, duration, err
}

func (r *Repository) InsertGymSubscriptionReturningID(ctx context.Context, db txtx.DBTX, args ...any) error {
	return db.QueryRow(ctx, models.QueryInsertGymSubscriptionReturningID, args[:len(args)-1]...).Scan(args[len(args)-1])
}

func (r *Repository) InsertGymSubscriptionFreePayment(ctx context.Context, db txtx.DBTX, args ...any) error {
	_, err := db.Exec(ctx, models.QueryInsertGymSubscriptionFreePayment, args...)
	return err
}

func (r *Repository) LogAudit(ctx context.Context, db txtx.DBTX, gymID int64, userID *int64, action, entityType string, entityID *int64, payload any) error {
	data, _ := json.Marshal(payload)
	_, err := db.Exec(ctx, models.QueryAuditLogInsert,
		token.GeneratePublicID(), gymID, userID, action, entityType, entityID, data)
	return err
}

func (r *Repository) Profile(ctx context.Context, userID, gymID int64) ([]map[string]any, error) {
	return r.queryMaps(ctx, models.QueryProfile, userID, gymID)
}

func (r *Repository) queryMaps(ctx context.Context, query string, args ...any) ([]map[string]any, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return scanMaps(rows)
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

func nullString(v any) sql.NullString {
	if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
		return sql.NullString{String: s, Valid: true}
	}
	return sql.NullString{}
}
