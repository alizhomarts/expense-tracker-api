package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateTransactionRequest struct {
	CategoryID  uuid.UUID `json:"category_id"`
	Type        string    `json:"type"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
}

type UpdateTransactionRequest struct {
	CategoryID  uuid.UUID `json:"category_id"`
	Type        string    `json:"type"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
}

type TransactionResponse struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	CategoryID  uuid.UUID `json:"category_id"`
	Type        string    `json:"type"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}
