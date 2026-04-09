package handler

import (
	"errors"
	"expense-tracker-api/internal/apperror"
	"expense-tracker-api/internal/dto"
	"expense-tracker-api/internal/service"
	"expense-tracker-api/logger"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"net/http"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c echo.Context) error {
	logger.Log.Info("user registration started")

	var req dto.RegisterRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": apperror.InvalidRequestBody.Error(),
		})
	}

	resp, err := h.authService.Register(c.Request().Context(), &req)
	if err != nil {
		logger.Log.WithError(err).Error("failed to register user")
		switch {
		case errors.Is(err, apperror.ErrUserAlreadyExists):
			return c.JSON(http.StatusConflict, map[string]string{
				"error": apperror.ErrUserAlreadyExists.Error(),
			})
		default:
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": apperror.InternalServer.Error(),
			})
		}
	}

	logger.Log.WithFields(logrus.Fields{
		"email": req.Email,
	}).Info("user registered successfully")

	return c.JSON(http.StatusCreated, resp)
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req dto.LoginRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": apperror.InvalidRequestBody.Error(),
		})
	}

	resp, err := h.authService.Login(c.Request().Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, apperror.ErrInvalidCredentials):
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": apperror.Unauthorized.Error(),
			})
		default:
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": apperror.InternalServer.Error(),
			})
		}
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Refresh(c echo.Context) error {
	var req dto.RefreshTokenRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": apperror.InvalidRequestBody.Error(),
		})
	}

	resp, err := h.authService.RefreshToken(c.Request().Context(), req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, apperror.ErrInvalidToken), errors.Is(err, apperror.ErrInvalidTokenType):
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": apperror.Unauthorized.Error(),
			})
		default:
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": apperror.InternalServer.Error(),
			})
		}
	}

	return c.JSON(http.StatusOK, resp)
}
