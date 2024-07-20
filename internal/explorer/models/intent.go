package models

import (
	"time"

	"github.com/google/uuid"
)

// Intent is a helper entity for managing remote workload objectives
type Intent struct {
	ID         uuid.UUID `json:"id"`
	Repository string    `json:"repository"`
	Since      time.Time `json:"since"`
	CreatedAt  time.Time `json:"created_at"`
	IsActive   bool      `json:"is_active"`
}

type IntentUpdate struct {
	ID       uuid.UUID
	IsActive bool       `json:"is_active"`
	Since    *time.Time `json:"since"`
}
