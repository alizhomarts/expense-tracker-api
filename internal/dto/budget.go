package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateBudgetRequest struct {
	CategoryID uuid.UUID `json:"category_id"`
	Amount     float64   `json:"amount"`
	Month      string    `json:"month"`
}

type BudgetResponse struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	CategoryID   uuid.UUID `json:"category_id"`
	CategoryName string    `json:"category_name,omitempty"`
	Amount       float64   `json:"amount"`
	Month        string    `json:"month"`
	CreatedAt    time.Time `json:"created_at"`
}

type BudgetStatusResponse struct {
	BudgetID     uuid.UUID `json:"budget_id"`
	CategoryID   uuid.UUID `json:"category_id"`
	CategoryName string    `json:"category_name"`
	Month        string    `json:"month"`
	BudgetAmount float64   `json:"budget_amount"`
	SpentAmount  float64   `json:"spent_amount"`
	Remaining    float64   `json:"remaining"`
	IsExceeded   bool      `json:"is_exceeded"`
	UsagePercent float64   `json:"usage_percent"`
}
