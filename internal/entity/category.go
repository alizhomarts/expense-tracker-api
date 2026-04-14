package entity

import (
	"github.com/google/uuid"
	"time"
)

type Category struct {
	ID        uuid.UUID
	UserID    *uuid.UUID
	Name      string
	Type      TransactionType
	CreatedAt time.Time
}
