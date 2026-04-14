package service

import (
	"context"
	"fmt"
	"regexp"

	"expense-tracker-api/internal/apperror"
	"expense-tracker-api/internal/dto"
	"expense-tracker-api/internal/entity"
	"expense-tracker-api/internal/repository"

	"github.com/google/uuid"
)

type BudgetService struct {
	budgetRepo repository.BudgetRepository
}

func NewBudgetService(budgetRepo repository.BudgetRepository) *BudgetService {
	return &BudgetService{
		budgetRepo: budgetRepo,
	}
}

func (s *BudgetService) Create(ctx context.Context, userID uuid.UUID, req *dto.CreateBudgetRequest) (*dto.BudgetResponse, error) {
	if req == nil {
		return nil, apperror.InvalidRequestBody
	}
	if userID == uuid.Nil {
		return nil, apperror.ErrUserRequired
	}
	if req.CategoryID == uuid.Nil {
		return nil, apperror.ErrIDRequired
	}
	if req.Amount <= 0 {
		return nil, apperror.ErrInvalidBudgetAmount
	}

	matched, _ := regexp.MatchString(`^\d{4}-\d{2}$`, req.Month)
	if !matched {
		return nil, apperror.ErrInvalidMonth
	}

	budget := &entity.Budget{
		UserID:     userID,
		CategoryID: req.CategoryID,
		Amount:     req.Amount,
		Month:      req.Month,
	}

	if err := s.budgetRepo.Create(ctx, budget); err != nil {
		return nil, fmt.Errorf("create budget: %w", err)
	}

	return &dto.BudgetResponse{
		ID:         budget.ID,
		UserID:     budget.UserID,
		CategoryID: budget.CategoryID,
		Amount:     budget.Amount,
		Month:      budget.Month,
		CreatedAt:  budget.CreatedAt,
	}, nil
}

func (s *BudgetService) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*dto.BudgetResponse, error) {
	if userID == uuid.Nil {
		return nil, apperror.ErrUserRequired
	}

	resp, err := s.budgetRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list budgets: %w", err)
	}

	return resp, nil
}

func (s *BudgetService) GetStatus(ctx context.Context, userID uuid.UUID, month string) ([]*dto.BudgetStatusResponse, error) {
	if userID == uuid.Nil {
		return nil, apperror.ErrUserRequired
	}

	matched, _ := regexp.MatchString(`^\d{4}-\d{2}$`, month)
	if !matched {
		return nil, apperror.ErrInvalidMonth
	}

	resp, err := s.budgetRepo.GetStatusByUserIDAndMonth(ctx, userID, month)
	if err != nil {
		return nil, fmt.Errorf("get budget status: %w", err)
	}

	return resp, nil
}
