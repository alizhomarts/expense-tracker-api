package service

import (
	"context"
	"fmt"
	"regexp"

	"expense-tracker-api/internal/apperror"
	"expense-tracker-api/internal/dto"
	"expense-tracker-api/internal/repository"

	"github.com/google/uuid"
)

type SummaryService struct {
	summaryRepo repository.SummaryRepository
}

func NewSummaryService(summaryRepo repository.SummaryRepository) *SummaryService {
	return &SummaryService{
		summaryRepo: summaryRepo,
	}
}

func (s *SummaryService) GetSummary(ctx context.Context, userID uuid.UUID) (*dto.SummaryResponse, error) {
	if userID == uuid.Nil {
		return nil, apperror.ErrUserRequired
	}

	resp, err := s.summaryRepo.GetSummary(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get summary: %w", err)
	}

	return resp, nil
}

func (s *SummaryService) GetCategorySummary(ctx context.Context, userID uuid.UUID) ([]*dto.CategorySummaryResponse, error) {
	if userID == uuid.Nil {
		return nil, apperror.ErrUserRequired
	}

	resp, err := s.summaryRepo.GetCategorySummary(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get category summary: %w", err)
	}

	return resp, nil
}

func (s *SummaryService) GetMonthlySummary(ctx context.Context, userID uuid.UUID, month string) (*dto.MonthlySummaryResponse, error) {
	if userID == uuid.Nil {
		return nil, apperror.ErrUserRequired
	}

	matched, _ := regexp.MatchString(`^\d{4}-\d{2}$`, month)
	if !matched {
		return nil, apperror.ErrInvalidMonth
	}

	resp, err := s.summaryRepo.GetMonthlySummary(ctx, userID, month)
	if err != nil {
		return nil, fmt.Errorf("get monthly summary: %w", err)
	}

	return resp, nil
}
