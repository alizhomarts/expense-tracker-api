package apperror

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrPasswordRequired   = errors.New("password required")
	ErrEmailRequired      = errors.New("email required")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrIDRequired         = errors.New("id is required")
	ErrUserRequired       = errors.New("user is required")
	ErrInvalidCredentials = errors.New("invalid credentials")

	Unauthorized       = errors.New("unauthorized")
	InvalidRequestBody = errors.New("invalid request body")
	InternalServer     = errors.New("internal server error")
	InvalidID          = errors.New("invalid id")

	ErrRefreshTokenRequired = errors.New("refresh token is required")
	ErrInvalidRefreshToken  = errors.New("invalid or expired refresh token")
	ErrMissingAuthHeader    = errors.New("missing authorization header")
	ErrInvalidAuthHeader    = errors.New("invalid authorization header")
	ErrInvalidToken         = errors.New("invalid token")
	ErrInvalidTokenType     = errors.New("invalid token type")
)
