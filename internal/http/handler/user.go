package handler

import (
	"errors"
	"net/http"

	"expense-tracker-api/internal/apperror"
	mymiddleware "expense-tracker-api/internal/http/middleware"
	"expense-tracker-api/internal/logger"
	"expense-tracker-api/internal/service"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

// GetMe godoc
// @Summary Get current user
// @Tags users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.UserResponse
// @Failure 401 {object} map[string]string
// @Router /users/me [get]
func (h *UserHandler) GetMe(c echo.Context) error {
	userID, err := mymiddleware.GetUserID(c)
	if err != nil {
		logger.Log.WithError(err).Error("get user id from context failed")

		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": apperror.Unauthorized.Error(),
		})
	}

	logger.Log.WithFields(logrus.Fields{
		"user_id": userID,
	}).Info("get current user request")

	resp, err := h.service.GetByID(c.Request().Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, apperror.ErrUserRequired):
			logger.Log.WithFields(logrus.Fields{
				"user_id": userID,
				"error":   err.Error(),
			}).Warn("get current user validation failed")

			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
		case errors.Is(err, apperror.ErrUserNotFound):
			logger.Log.WithFields(logrus.Fields{
				"user_id": userID,
			}).Warn("current user not found")

			return c.JSON(http.StatusNotFound, map[string]string{
				"error": apperror.ErrUserNotFound.Error(),
			})
		default:
			logger.Log.WithFields(logrus.Fields{
				"user_id": userID,
				"error":   err.Error(),
			}).Error("get current user failed")

			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": apperror.InternalServer.Error(),
			})
		}
	}

	logger.Log.WithFields(logrus.Fields{
		"user_id": userID,
		"email":   resp.Email,
	}).Info("current user returned successfully")

	return c.JSON(http.StatusOK, resp)
}
