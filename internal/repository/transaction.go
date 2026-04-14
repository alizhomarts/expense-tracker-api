package repository

import (
	"context"
	"errors"
	"expense-tracker-api/internal/apperror"
	"expense-tracker-api/internal/entity"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type (
	TransactionRepository interface {
		Create(ctx context.Context, transaction *entity.Transaction) error
		GetByID(ctx context.Context, id, userID uuid.UUID) (*entity.Transaction, error)
		ListByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Transaction, error)
		Update(ctx context.Context, transaction *entity.Transaction) error
		Delete(ctx context.Context, id, userID uuid.UUID) error
	}

	transactionRepository struct {
		db *pgxpool.Pool
	}
)

func NewTransactionRepository(db *pgxpool.Pool) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(ctx context.Context, transaction *entity.Transaction) error {
	query := `
		INSERT INTO transactions (user_id, category_id, type, amount, description)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(ctx, query,
		transaction.UserID,
		transaction.CategoryID,
		transaction.Type,
		transaction.Amount,
		transaction.Description,
	).Scan(&transaction.ID, &transaction.CreatedAt)
	if err != nil {
		return fmt.Errorf("create transaction: %w", err)
	}

	return nil
}

func (r *transactionRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*entity.Transaction, error) {
	var transaction entity.Transaction
	query := `
		select id, user_id, category_id, type, amount, description, created_at
		from transactions
		where id = $1 and user_id = $2
	`

	err := r.db.QueryRow(ctx, query, id, userID).Scan(
		&transaction.ID,
		&transaction.UserID,
		&transaction.CategoryID,
		&transaction.Type,
		&transaction.Amount,
		&transaction.Description,
		&transaction.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrTransactionNotFound
		}
		return nil, fmt.Errorf("get transaction by id: %w", err)
	}

	return &transaction, nil
}

func (r *transactionRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Transaction, error) {
	query := `
		SELECT id, user_id, category_id, type, amount, description, created_at
		FROM transactions
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list transactions by user id: %w", err)
	}
	defer rows.Close()

	var transactions []*entity.Transaction
	for rows.Next() {
		var transaction entity.Transaction
		err := rows.Scan(
			&transaction.ID,
			&transaction.UserID,
			&transaction.CategoryID,
			&transaction.Type,
			&transaction.Amount,
			&transaction.Description,
			&transaction.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("list transactions by user id: %w", err)
		}

		transactions = append(transactions, &transaction)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return transactions, nil
}

func (r *transactionRepository) Update(ctx context.Context, transaction *entity.Transaction) error {
	query := `
		UPDATE transactions
		SET category_id = $1,
		    type = $2,
		    amount = $3,
		    description = $4
		WHERE id = $5 AND user_id = $6
	`

	cmdTag, err := r.db.Exec(ctx, query,
		transaction.CategoryID,
		transaction.Type,
		transaction.Amount,
		transaction.Description,
		transaction.ID,
		transaction.UserID,
	)
	if err != nil {
		return fmt.Errorf("update transaction: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return apperror.ErrTransactionNotFound
	}

	return nil
}

func (r *transactionRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	query := `
		DELETE FROM transactions
		WHERE id = $1 AND user_id = $2
	`

	cmdTag, err := r.db.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("delete transaction: %w", err)

	}

	if cmdTag.RowsAffected() == 0 {
		return apperror.ErrTransactionNotFound
	}

	return nil
}
