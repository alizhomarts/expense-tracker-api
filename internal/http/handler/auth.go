package handler

import (
	"errors"
	"expense-tracker-api/internal/apperror"
	"expense-tracker-api/internal/dto"
	"expense-tracker-api/internal/logger"
	"expense-tracker-api/internal/service"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"net/http"
)

type AuthHandler struct {
	service *service.AuthService
}

func NewAuthHandler(service *service.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

// Register godoc
// @Summary Register user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Register request"
// @Success 201 {object} dto.UserResponse
// @Failure 400 {object} map[string]string
// @Router /auth/register [post]
func (h *AuthHandler) Register(c echo.Context) error {
	var req dto.RegisterRequest

	if err := c.Bind(&req); err != nil {
		logger.Log.WithError(err).Error("bind register request failed")

		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": apperror.InvalidRequestBody.Error(),
		})
	}

	logger.Log.WithFields(logrus.Fields{
		"email":      req.Email,
		"first_name": req.FirstName,
		"last_name":  req.LastName,
	}).Info("register request received")

	resp, err := h.service.Register(c.Request().Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, apperror.ErrUserAlreadyExists):
			logger.Log.WithFields(logrus.Fields{
				"email": req.Email,
				"error": err.Error(),
			}).Warn("register failed: user already exists")

			return c.JSON(http.StatusConflict, map[string]string{
				"error": apperror.ErrUserAlreadyExists.Error(),
			})
		default:
			logger.Log.WithFields(logrus.Fields{
				"email": req.Email,
				"error": err.Error(),
			}).Error("register failed")

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

// Login godoc
// @Summary Login user
// @Description Authenticate user and return access and refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login request"
// @Success 200 {object} dto.AuthResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/login [post]
func (h *AuthHandler) Login(c echo.Context) error {
	var req dto.LoginRequest

	if err := c.Bind(&req); err != nil {
		logger.Log.WithError(err).Error("bind login request failed")

		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": apperror.InvalidRequestBody.Error(),
		})
	}

	logger.Log.WithFields(logrus.Fields{
		"email": req.Email,
	}).Info("login request received")

	resp, err := h.service.Login(c.Request().Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, apperror.ErrInvalidCredentials):
			logger.Log.WithFields(logrus.Fields{
				"email": req.Email,
			}).Warn("login failed: invalid credentials")

			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": apperror.Unauthorized.Error(),
			})
		default:
			logger.Log.WithFields(logrus.Fields{
				"email": req.Email,
				"error": err.Error(),
			}).Error("login failed")

			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": apperror.InternalServer.Error(),
			})
		}
	}

	logger.Log.WithFields(logrus.Fields{
		"email": req.Email,
	}).Info("user logged in successfully")

	return c.JSON(http.StatusOK, resp)
}

// Refresh godoc
// @Summary Refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Refresh request"
// @Success 200 {object} dto.AuthResponse
// @Failure 401 {object} map[string]string
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(c echo.Context) error {
	var req dto.RefreshTokenRequest

	if err := c.Bind(&req); err != nil {
		logger.Log.WithError(err).Error("bind refresh token request failed")

		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": apperror.InvalidRequestBody.Error(),
		})
	}

	logger.Log.Info("refresh token request received")

	resp, err := h.service.RefreshToken(c.Request().Context(), req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, apperror.ErrInvalidToken),
			errors.Is(err, apperror.ErrInvalidTokenType):
			logger.Log.WithFields(logrus.Fields{
				"error": err.Error(),
			}).Warn("refresh token failed: invalid token")

			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": apperror.Unauthorized.Error(),
			})
		default:
			logger.Log.WithFields(logrus.Fields{
				"error": err.Error(),
			}).Error("refresh token failed")

			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": apperror.InternalServer.Error(),
			})
		}
	}

	logger.Log.Info("tokens refreshed successfully")

	return c.JSON(http.StatusOK, resp)
}
