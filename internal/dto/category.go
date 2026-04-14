package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateCategoryRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type CategoryResponse struct {
	ID        uuid.UUID  `json:"id"`
	UserID    *uuid.UUID `json:"user_id,omitempty"`
	Name      string     `json:"name"`
	Type      string     `json:"type"`
	CreatedAt time.Time  `json:"created_at"`
}
