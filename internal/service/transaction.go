package service

import (
	"context"
	"expense-tracker-api/internal/apperror"
	"expense-tracker-api/internal/dto"
	"expense-tracker-api/internal/entity"
	"expense-tracker-api/internal/repository"
	"fmt"
	"github.com/google/uuid"
	"strings"
)

type TransactionService struct {
	repo repository.TransactionRepository
}

func NewTransactionService(repo repository.TransactionRepository) *TransactionService {
	return &TransactionService{repo: repo}
}

func (s *TransactionService) Create(ctx context.Context, userID uuid.UUID, req *dto.CreateTransactionRequest) (*dto.TransactionResponse, error) {
	if userID == uuid.Nil {
		return nil, apperror.ErrUserRequired
	}
	if req == nil {
		return nil, apperror.InvalidRequestBody
	}
	if err := s.validateTransactionInput(req.Type, req.Amount, req.Description); err != nil {
		return nil, err
	}

	transactionType := strings.TrimSpace(req.Type)
	description := strings.TrimSpace(req.Description)

	transaction := &entity.Transaction{
		UserID:      userID,
		CategoryID:  req.CategoryID,
		Type:        entity.TransactionType(transactionType),
		Amount:      req.Amount,
		Description: description,
	}

	if err := s.repo.Create(ctx, transaction); err != nil {
		return nil, fmt.Errorf("create transaction: %w", err)
	}

	return toTransactionResponse(transaction), nil
}

func (s *TransactionService) GetByID(ctx context.Context, id, userID uuid.UUID) (*dto.TransactionResponse, error) {
	if id == uuid.Nil {
		return nil, apperror.ErrIDRequired
	}
	if userID == uuid.Nil {
		return nil, apperror.ErrUserRequired
	}

	transaction, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("get transaction by id: %w", err)
	}

	return toTransactionResponse(transaction), nil
}

func (s *TransactionService) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*dto.TransactionResponse, error) {
	if userID == uuid.Nil {
		return nil, apperror.ErrUserRequired
	}

	transactions, err := s.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list transactions by user id: %w", err)
	}

	return toTransactionResponses(transactions), nil
}

func (s *TransactionService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	if id == uuid.Nil {
		return apperror.ErrIDRequired
	}
	if userID == uuid.Nil {
		return apperror.ErrUserRequired
	}

	err := s.repo.Delete(ctx, id, userID)
	if err != nil {
		return fmt.Errorf("delete transaction: %w", err)
	}

	return nil
}

func (s *TransactionService) Update(ctx context.Context, id, userID uuid.UUID, req *dto.UpdateTransactionRequest) (*dto.TransactionResponse, error) {
	if id == uuid.Nil {
		return nil, apperror.ErrIDRequired
	}
	if userID == uuid.Nil {
		return nil, apperror.ErrUserRequired
	}
	if req == nil {
		return nil, apperror.InvalidRequestBody
	}

	if err := s.validateTransactionInput(req.Type, req.Amount, req.Description); err != nil {
		return nil, err
	}

	transaction, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("get transaction before update: %w", err)
	}

	transactionType := strings.TrimSpace(req.Type)
	description := strings.TrimSpace(req.Description)

	transaction.CategoryID = req.CategoryID
	transaction.Type = entity.TransactionType(transactionType)
	transaction.Amount = req.Amount
	transaction.Description = description

	if err := s.repo.Update(ctx, transaction); err != nil {
		return nil, fmt.Errorf("update transaction: %w", err)
	}

	return toTransactionResponse(transaction), nil
}

func (s *TransactionService) validateTransactionInput(transactionType string, amount float64, description string) error {
	transactionType = strings.TrimSpace(transactionType)
	description = strings.TrimSpace(description)

	if transactionType != string(entity.TransactionTypeIncome) &&
		transactionType != string(entity.TransactionTypeExpense) {
		return apperror.ErrInvalidTransactionType
	}

	if amount <= 0 {
		return apperror.ErrInvalidAmount
	}

	if len(description) > 255 {
		return apperror.ErrInvalidDescription
	}

	return nil
}

func toTransactionResponse(transaction *entity.Transaction) *dto.TransactionResponse {
	return &dto.TransactionResponse{
		ID:          transaction.ID,
		UserID:      transaction.UserID,
		CategoryID:  transaction.CategoryID,
		Type:        string(transaction.Type),
		Amount:      transaction.Amount,
		Description: transaction.Description,
		CreatedAt:   transaction.CreatedAt,
	}
}

func toTransactionResponses(transactions []*entity.Transaction) []*dto.TransactionResponse {
	responses := make([]*dto.TransactionResponse, 0, len(transactions))

	for _, transaction := range transactions {
		responses = append(responses, toTransactionResponse(transaction))
	}

	return responses
}
