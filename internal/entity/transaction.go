package entity

import (
	"github.com/google/uuid"
	"time"
)

type TransactionType string

const (
	TransactionTypeIncome  TransactionType = "income"
	TransactionTypeExpense TransactionType = "expense"
)

type Transaction struct {
	ID          uuid.UUID       `json:"id"`
	UserID      uuid.UUID       `json:"user_id"`
	Category    uuid.UUID       `json:"category_id"`
	Type        TransactionType `json:"transaction_type"`
	Amount      float64         `json:"amount"`
	Description string          `json:"description"`
	CreatedAt   time.Time       `json:"created_at"`
}
