package entity

import (
	"github.com/google/uuid"
	"time"
)

type Budget struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	CategoryID uuid.UUID `json:"category_id"`
	Amount     float64   `json:"amount"`
	Month      string    `json:"month"`
	CreatedAt  time.Time `json:"created_at"`
}
