package repository

import (
	"context"
	"errors"
	"fmt"

	"expense-tracker-api/internal/apperror"
	"expense-tracker-api/internal/entity"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type (
	CategoryRepository interface {
		Create(ctx context.Context, category *entity.Category) error
		ListByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Category, error)
		GetByNameAndType(ctx context.Context, userID uuid.UUID, name string, categoryType entity.TransactionType) (*entity.Category, error)
	}

	categoryRepository struct {
		db *pgxpool.Pool
	}
)

func NewCategoryRepository(db *pgxpool.Pool) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(ctx context.Context, category *entity.Category) error {
	query := `
		INSERT INTO categories (user_id, name, type)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(ctx, query,
		category.UserID,
		category.Name,
		category.Type,
	).Scan(&category.ID, &category.CreatedAt)
	if err != nil {
		return fmt.Errorf("create category: %w", err)
	}

	return nil
}

func (r *categoryRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Category, error) {
	query := `
		SELECT id, user_id, name, type, created_at
		FROM categories
		WHERE user_id = $1 OR user_id IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list categories by user id: %w", err)
	}
	defer rows.Close()

	var categories []*entity.Category

	for rows.Next() {
		var category entity.Category

		if err := rows.Scan(
			&category.ID,
			&category.UserID,
			&category.Name,
			&category.Type,
			&category.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}

		categories = append(categories, &category)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return categories, nil
}

func (r *categoryRepository) GetByNameAndType(ctx context.Context, userID uuid.UUID, name string, categoryType entity.TransactionType) (*entity.Category, error) {
	var category entity.Category

	query := `
		SELECT id, user_id, name, type, created_at
		FROM categories
		WHERE (user_id = $1 OR user_id IS NULL)
		  AND name = $2
		  AND type = $3
		ORDER BY user_id DESC NULLS LAST
		LIMIT 1
	`

	err := r.db.QueryRow(ctx, query, userID, name, categoryType).Scan(
		&category.ID,
		&category.UserID,
		&category.Name,
		&category.Type,
		&category.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrCategoryNotFound
		}
		return nil, fmt.Errorf("get category by name and type: %w", err)
	}

	return &category, nil
}
