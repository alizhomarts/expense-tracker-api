package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"expense-tracker-api/internal/apperror"
	"expense-tracker-api/internal/dto"
	"expense-tracker-api/internal/entity"
	"expense-tracker-api/internal/repository"

	"github.com/google/uuid"
)

type CategoryService struct {
	repo repository.CategoryRepository
}

func NewCategoryService(repo repository.CategoryRepository) *CategoryService {
	return &CategoryService{
		repo: repo,
	}
}

func (s *CategoryService) Create(ctx context.Context, userID uuid.UUID, req *dto.CreateCategoryRequest) (*dto.CategoryResponse, error) {
	if req == nil {
		return nil, apperror.InvalidRequestBody
	}
	if userID == uuid.Nil {
		return nil, apperror.ErrUserRequired
	}

	name := strings.TrimSpace(req.Name)
	categoryType := strings.TrimSpace(req.Type)

	if name == "" {
		return nil, apperror.ErrInvalidCategoryName
	}

	if categoryType != string(entity.TransactionTypeIncome) &&
		categoryType != string(entity.TransactionTypeExpense) {
		return nil, apperror.ErrInvalidTransactionType
	}

	category := &entity.Category{
		UserID: &userID,
		Name:   name,
		Type:   entity.TransactionType(categoryType),
	}

	if err := s.repo.Create(ctx, category); err != nil {
		return nil, fmt.Errorf("create category: %w", err)
	}

	return toCategoryResponse(category), nil
}

func (s *CategoryService) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*dto.CategoryResponse, error) {
	if userID == uuid.Nil {
		return nil, apperror.ErrUserRequired
	}

	categories, err := s.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list categories by user id: %w", err)
	}

	return toCategoryResponses(categories), nil
}

func (s *CategoryService) GetOrCreate(
	ctx context.Context,
	userID uuid.UUID,
	name string,
	categoryType entity.TransactionType,
) (*entity.Category, error) {
	if userID == uuid.Nil {
		return nil, apperror.ErrUserRequired
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return nil, apperror.ErrInvalidCategoryName
	}

	if categoryType != entity.TransactionTypeExpense &&
		categoryType != entity.TransactionTypeIncome {
		return nil, apperror.ErrInvalidTransactionType
	}

	category, err := s.repo.GetByNameAndType(ctx, userID, name, categoryType)
	if err == nil {
		return category, nil
	}

	if !errors.Is(err, apperror.ErrCategoryNotFound) {
		return nil, fmt.Errorf("get category by name and type: %w", err)
	}

	newCategory := &entity.Category{
		UserID: &userID,
		Name:   name,
		Type:   categoryType,
	}

	if err := s.repo.Create(ctx, newCategory); err != nil {
		return nil, fmt.Errorf("create category: %w", err)
	}

	return newCategory, nil
}

func toCategoryResponse(category *entity.Category) *dto.CategoryResponse {
	return &dto.CategoryResponse{
		ID:        category.ID,
		UserID:    category.UserID,
		Name:      category.Name,
		Type:      string(category.Type),
		CreatedAt: category.CreatedAt,
	}
}

func toCategoryResponses(categories []*entity.Category) []*dto.CategoryResponse {
	responses := make([]*dto.CategoryResponse, 0, len(categories))

	for _, category := range categories {
		responses = append(responses, toCategoryResponse(category))
	}

	return responses
}
