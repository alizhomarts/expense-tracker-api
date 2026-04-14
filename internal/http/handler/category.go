package handler

import (
	"errors"
	"net/http"

	"expense-tracker-api/internal/apperror"
	"expense-tracker-api/internal/dto"
	mymiddleware "expense-tracker-api/internal/http/middleware"
	"expense-tracker-api/internal/logger"
	"expense-tracker-api/internal/service"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type CategoryHandler struct {
	categoryService *service.CategoryService
}

func NewCategoryHandler(categoryService *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
	}
}

func (h *CategoryHandler) Create(c echo.Context) error {
	var req dto.CreateCategoryRequest

	if err := c.Bind(&req); err != nil {
		logger.Log.WithError(err).Error("bind create category request failed")

		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": apperror.InvalidRequestBody.Error(),
		})
	}

	userID, err := mymiddleware.GetUserID(c)
	if err != nil {
		logger.Log.WithError(err).Error("get user id from context failed")

		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": apperror.Unauthorized.Error(),
		})
	}

	logger.Log.WithFields(logrus.Fields{
		"user_id": userID,
		"name":    req.Name,
		"type":    req.Type,
	}).Info("create category request")

	resp, err := h.categoryService.Create(c.Request().Context(), userID, &req)
	if err != nil {
		switch {
		case errors.Is(err, apperror.ErrUserRequired),
			errors.Is(err, apperror.ErrInvalidCategoryName),
			errors.Is(err, apperror.ErrInvalidTransactionType),
			errors.Is(err, apperror.InvalidRequestBody):
			logger.Log.WithFields(logrus.Fields{
				"user_id": userID,
				"error":   err.Error(),
			}).Warn("create category validation failed")

			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
		default:
			logger.Log.WithFields(logrus.Fields{
				"user_id": userID,
				"error":   err.Error(),
			}).Error("create category failed")

			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": apperror.InternalServer.Error(),
			})
		}
	}

	return c.JSON(http.StatusCreated, resp)
}

func (h *CategoryHandler) List(c echo.Context) error {
	userID, err := mymiddleware.GetUserID(c)
	if err != nil {
		logger.Log.WithError(err).Error("get user id from context failed")

		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": apperror.Unauthorized.Error(),
		})
	}

	logger.Log.WithFields(logrus.Fields{
		"user_id": userID,
	}).Info("list categories request")

	resp, err := h.categoryService.ListByUserID(c.Request().Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, apperror.ErrUserRequired):
			logger.Log.WithFields(logrus.Fields{
				"user_id": userID,
				"error":   err.Error(),
			}).Warn("list categories validation failed")

			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
		default:
			logger.Log.WithFields(logrus.Fields{
				"user_id": userID,
				"error":   err.Error(),
			}).Error("list categories failed")

			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": apperror.InternalServer.Error(),
			})
		}
	}

	return c.JSON(http.StatusOK, resp)
}
