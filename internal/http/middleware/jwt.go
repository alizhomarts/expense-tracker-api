package middleware

import (
	"errors"
	"expense-tracker-api/internal/apperror"
	"expense-tracker-api/internal/logger"
	"expense-tracker-api/internal/service"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
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
		method := c.Request().Method
		path := c.Request().URL.Path
		ip := c.RealIP()

		logger.Log.WithFields(logrus.Fields{
			"method": method,
			"path":   path,
			"ip":     ip,
		}).Info("jwt middleware started")

		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			logger.Log.WithFields(logrus.Fields{
				"method": method,
				"path":   path,
				"ip":     ip,
			}).Warn("missing authorization header")

			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": apperror.ErrMissingAuthHeader.Error(),
			})
		}

		parts := strings.Fields(authHeader)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			logger.Log.WithFields(logrus.Fields{
				"method":             method,
				"path":               path,
				"ip":                 ip,
				"auth_header_length": len(authHeader),
			}).Warn("invalid authorization header format")

			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": apperror.ErrInvalidAuthHeader.Error(),
			})
		}

		tokenString := parts[1]

		logger.Log.WithFields(logrus.Fields{
			"method":       method,
			"path":         path,
			"ip":           ip,
			"token_length": len(tokenString),
		}).Info("authorization header parsed")

		userID, err := m.authService.ParseToken(tokenString, service.TokenTypeAccess)
		if err != nil {
			level := logger.Log.WithFields(logrus.Fields{
				"method": method,
				"path":   path,
				"ip":     ip,
				"error":  err.Error(),
			})

			if errors.Is(err, apperror.ErrInvalidToken) || errors.Is(err, apperror.ErrInvalidTokenType) {
				level.Warn("access token validation failed")
			} else {
				level.Error("unexpected jwt parse error")
			}

			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": apperror.Unauthorized.Error(),
			})
		}

		c.Set(ContextUserIDKey, userID)

		logger.Log.WithFields(logrus.Fields{
			"method":  method,
			"path":    path,
			"ip":      ip,
			"user_id": userID,
		}).Info("jwt middleware passed successfully")

		return next(c)
	}
}

func GetUserID(c echo.Context) (uuid.UUID, error) {
	userIDValue := c.Get(ContextUserIDKey)

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		return uuid.Nil, apperror.ErrInvalidToken
	}

	return userID, nil
}
