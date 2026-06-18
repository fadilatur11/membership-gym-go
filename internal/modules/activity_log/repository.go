package activity_log

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"membership-gym/internal/models"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(
	ctx context.Context,
	gymID any,
	userID any,
	requestPayload any,
	response any,
	curl string,
	statusCode int,
	status string,
	responseTime int64,
) error {
	_, err := r.db.Exec(
		ctx,
		models.QueryActivityLogInsert,
		gymID,
		userID,
		requestPayload,
		response,
		curl,
		statusCode,
		status,
		responseTime,
	)
	return err
}
