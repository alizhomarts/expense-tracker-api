package repository

import (
	"context"
	"fmt"

	"expense-tracker-api/internal/dto"
	"expense-tracker-api/internal/entity"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BudgetRepository interface {
	Create(ctx context.Context, budget *entity.Budget) error
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*dto.BudgetResponse, error)
	GetStatusByUserIDAndMonth(ctx context.Context, userID uuid.UUID, month string) ([]*dto.BudgetStatusResponse, error)
}

type budgetRepository struct {
	db *pgxpool.Pool
}

func NewBudgetRepository(db *pgxpool.Pool) BudgetRepository {
	return &budgetRepository{db: db}
}

func (r *budgetRepository) Create(ctx context.Context, budget *entity.Budget) error {
	query := `
		INSERT INTO budgets (user_id, category_id, amount, month)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(ctx, query,
		budget.UserID,
		budget.CategoryID,
		budget.Amount,
		budget.Month,
	).Scan(&budget.ID, &budget.CreatedAt)
	if err != nil {
		return fmt.Errorf("create budget: %w", err)
	}

	return nil
}

func (r *budgetRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*dto.BudgetResponse, error) {
	query := `
		SELECT
			b.id,
			b.user_id,
			b.category_id,
			c.name,
			b.amount,
			b.month,
			b.created_at
		FROM budgets b
		JOIN categories c ON c.id = b.category_id
		WHERE b.user_id = $1
		ORDER BY b.created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list budgets by user id: %w", err)
	}
	defer rows.Close()

	var budgets []*dto.BudgetResponse

	for rows.Next() {
		var item dto.BudgetResponse
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.CategoryID,
			&item.CategoryName,
			&item.Amount,
			&item.Month,
			&item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan budget: %w", err)
		}

		budgets = append(budgets, &item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("budget rows error: %w", err)
	}

	return budgets, nil
}

func (r *budgetRepository) GetStatusByUserIDAndMonth(ctx context.Context, userID uuid.UUID, month string) ([]*dto.BudgetStatusResponse, error) {
	query := `
		SELECT
			b.id AS budget_id,
			b.category_id,
			c.name AS category_name,
			b.month,
			b.amount AS budget_amount,
			COALESCE(SUM(t.amount), 0) AS spent_amount
		FROM budgets b
		JOIN categories c ON c.id = b.category_id
		LEFT JOIN transactions t
			ON t.category_id = b.category_id
			AND t.user_id = b.user_id
			AND t.type = 'expense'
			AND TO_CHAR(t.created_at, 'YYYY-MM') = b.month
		WHERE b.user_id = $1
		  AND b.month = $2
		GROUP BY b.id, b.category_id, c.name, b.month, b.amount
		ORDER BY spent_amount DESC
	`

	rows, err := r.db.Query(ctx, query, userID, month)
	if err != nil {
		return nil, fmt.Errorf("get budget status: %w", err)
	}
	defer rows.Close()

	var result []*dto.BudgetStatusResponse

	for rows.Next() {
		var item dto.BudgetStatusResponse
		if err := rows.Scan(
			&item.BudgetID,
			&item.CategoryID,
			&item.CategoryName,
			&item.Month,
			&item.BudgetAmount,
			&item.SpentAmount,
		); err != nil {
			return nil, fmt.Errorf("scan budget status: %w", err)
		}

		item.Remaining = item.BudgetAmount - item.SpentAmount
		item.IsExceeded = item.SpentAmount > item.BudgetAmount

		if item.BudgetAmount > 0 {
			item.UsagePercent = (item.SpentAmount / item.BudgetAmount) * 100
		}

		result = append(result, &item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("budget status rows error: %w", err)
	}

	return result, nil
}
