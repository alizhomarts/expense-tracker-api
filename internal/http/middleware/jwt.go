package middleware

import (
	"errors"
	"expense-tracker-api/internal/apperror"
	"expense-tracker-api/internal/service"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
)

const ContextUserIDKey = "user_id"

type JWTMiddleware struct {
	authService *service.AuthService
}

func NewJWTMiddleware(authService *service.AuthService) *JWTMiddleware {
	return &JWTMiddleware{authService: authService}
}

func (m *JWTMiddleware) Handle(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": apperror.Unauthorized.Error(),
			})
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": apperror.Unauthorized.Error(),
			})
		}

		tokenString := parts[1]

		userID, err := m.authService.ParseToken(tokenString, service.TokenTypeAccess)
		if err != nil {
			switch {
			case errors.Is(err, apperror.ErrInvalidToken), errors.Is(err, apperror.ErrInvalidTokenType):
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": apperror.Unauthorized.Error(),
				})
			default:
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": apperror.Unauthorized.Error(),
				})
			}
		}

		c.Set(ContextUserIDKey, userID)

		return next(c)
	}
}
