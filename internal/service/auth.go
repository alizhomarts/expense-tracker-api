package service

import (
	"context"
	"errors"
	"expense-tracker-api/internal/apperror"
	"expense-tracker-api/internal/dto"
	"expense-tracker-api/internal/entity"
	"expense-tracker-api/internal/repository"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"time"
)

const (
	accessTokenTTL   = 15 * time.Minute
	refreshTokenTTL  = 7 * 24 * time.Hour
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

type AuthService struct {
	repo          repository.UserRepository
	accessSecret  string
	refreshSecret string
}

func NewAuthService(repo repository.UserRepository, accessSecret, refreshSecret string) *AuthService {
	return &AuthService{
		repo:          repo,
		accessSecret:  accessSecret,
		refreshSecret: refreshSecret,
	}
}

func (s *AuthService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, error) {
	if req == nil {
		return nil, apperror.ErrUserRequired
	}
	if req.Email == "" {
		return nil, apperror.ErrEmailRequired
	}
	if req.Password == "" {
		return nil, apperror.ErrPasswordRequired
	}
	_, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, apperror.ErrUserNotFound) {
		return nil, fmt.Errorf("check user by email: %w", err)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &entity.User{
		Email:        req.Email,
		PasswordHash: string(passwordHash),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
	}

	err = s.repo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return s.generateTokenPair(user.ID)
}

func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, apperror.ErrUserNotFound) {
			return nil, apperror.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, apperror.ErrInvalidCredentials
	}

	return s.generateTokenPair(user.ID)
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*dto.AuthResponse, error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil, apperror.ErrRefreshTokenRequired
	}

	userID, err := s.ParseToken(refreshToken, TokenTypeRefresh)
	if err != nil {
		return nil, err
	}

	return s.generateTokenPair(userID)
}

func (s *AuthService) ParseToken(tokenString string, expectedType string) (uuid.UUID, error) {
	var secret string

	switch expectedType {
	case TokenTypeAccess:
		secret = s.accessSecret
	case TokenTypeRefresh:
		secret = s.refreshSecret
	default:
		return uuid.Nil, apperror.ErrInvalidTokenType
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if token.Method == nil || token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, apperror.ErrInvalidToken
		}
		return []byte(secret), nil
	})
	if err != nil {
		return uuid.Nil, apperror.ErrInvalidToken
	}

	if !token.Valid {
		return uuid.Nil, apperror.ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, apperror.ErrInvalidToken
	}

	claimType, ok := claims["type"].(string)
	if !ok {
		return uuid.Nil, apperror.ErrInvalidToken
	}

	if claimType != expectedType {
		return uuid.Nil, apperror.ErrInvalidTokenType
	}

	userIDValue, ok := claims["user_id"].(string)
	if !ok {
		return uuid.Nil, apperror.ErrInvalidToken
	}

	userID, err := uuid.Parse(userIDValue)
	if err != nil {
		return uuid.Nil, apperror.ErrInvalidToken
	}

	return userID, nil
}

func (s *AuthService) generateTokenPair(userID uuid.UUID) (*dto.AuthResponse, error) {
	accessToken, err := s.generateToken(userID, TokenTypeAccess, accessTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := s.generateToken(userID, TokenTypeRefresh, refreshTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) generateToken(userID uuid.UUID, tokenType string, ttl time.Duration) (string, error) {
	var secret string

	switch tokenType {
	case TokenTypeAccess:
		secret = s.accessSecret
	case TokenTypeRefresh:
		secret = s.refreshSecret
	default:
		return "", fmt.Errorf("invalid token type: %s", tokenType)
	}

	now := time.Now()

	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"type":    tokenType,
		"exp":     now.Add(ttl).Unix(),
		"iat":     now.Unix(),
		"nbf":     now.Unix(),
		"iss":     "expense-tracker-api",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return tokenString, nil
}
