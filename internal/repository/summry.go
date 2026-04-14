package repository

import (
	"context"
	"fmt"

	"expense-tracker-api/internal/dto"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SummaryRepository interface {
	GetSummary(ctx context.Context, userID uuid.UUID) (*dto.SummaryResponse, error)
	GetCategorySummary(ctx context.Context, userID uuid.UUID) ([]*dto.CategorySummaryResponse, error)
	GetMonthlySummary(ctx context.Context, userID uuid.UUID, month string) (*dto.MonthlySummaryResponse, error)
}

type summaryRepository struct {
	db *pgxpool.Pool
}

func NewSummaryRepository(db *pgxpool.Pool) SummaryRepository {
	return &summaryRepository{db: db}
}

func (r *summaryRepository) GetSummary(ctx context.Context, userID uuid.UUID) (*dto.SummaryResponse, error) {
	query := `
		SELECT
			COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0) AS total_income,
			COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0) AS total_expense
		FROM transactions
		WHERE user_id = $1
	`

	var resp dto.SummaryResponse

	err := r.db.QueryRow(ctx, query, userID).Scan(
		&resp.TotalIncome,
		&resp.TotalExpense,
	)
	if err != nil {
		return nil, fmt.Errorf("get summary: %w", err)
	}

	resp.Balance = resp.TotalIncome - resp.TotalExpense

	return &resp, nil
}

func (r *summaryRepository) GetCategorySummary(ctx context.Context, userID uuid.UUID) ([]*dto.CategorySummaryResponse, error) {
	query := `
		SELECT
			COALESCE(c.id::text, '') AS category_id,
			COALESCE(c.name, 'uncategorized') AS category_name,
			COALESCE(SUM(t.amount), 0) AS total
		FROM transactions t
		LEFT JOIN categories c ON c.id = t.category_id
		WHERE t.user_id = $1
		  AND t.type = 'expense'
		GROUP BY c.id, c.name
		ORDER BY total DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("get category summary: %w", err)
	}
	defer rows.Close()

	var result []*dto.CategorySummaryResponse

	for rows.Next() {
		var item dto.CategorySummaryResponse

		if err := rows.Scan(
			&item.CategoryID,
			&item.CategoryName,
			&item.Total,
		); err != nil {
			return nil, fmt.Errorf("scan category summary: %w", err)
		}

		result = append(result, &item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("category summary rows error: %w", err)
	}

	return result, nil
}

func (r *summaryRepository) GetMonthlySummary(ctx context.Context, userID uuid.UUID, month string) (*dto.MonthlySummaryResponse, error) {
	query := `
		SELECT
			COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0) AS total_income,
			COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0) AS total_expense
		FROM transactions
		WHERE user_id = $1
		  AND TO_CHAR(created_at, 'YYYY-MM') = $2
	`

	var resp dto.MonthlySummaryResponse
	resp.Month = month

	err := r.db.QueryRow(ctx, query, userID, month).Scan(
		&resp.TotalIncome,
		&resp.TotalExpense,
	)
	if err != nil {
		return nil, fmt.Errorf("get monthly summary: %w", err)
	}

	resp.Balance = resp.TotalIncome - resp.TotalExpense

	return &resp, nil
}
