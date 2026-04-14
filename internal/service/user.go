package service

import (
	"context"
	"fmt"
	"strings"

	"expense-tracker-api/internal/apperror"
	"expense-tracker-api/internal/dto"
	"expense-tracker-api/internal/repository"

	"github.com/google/uuid"
)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (s *UserService) GetByID(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error) {
	if userID == uuid.Nil {
		return nil, apperror.ErrUserRequired
	}

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return &dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: user.CreatedAt,
	}, nil
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*dto.UserResponse, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return nil, apperror.ErrEmailRequired
	}

	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return &dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: user.CreatedAt,
	}, nil
}
