package models

import (
	"time"

	"github.com/google/uuid"
)

type BaseModel struct {
	ID        int64
	PublicID  uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

type BaseTenantModel struct {
	BaseModel
	GymID int64
}
