package handler

import (
	"errors"
	"net/http"
	"strings"

	"expense-tracker-api/internal/apperror"
	"expense-tracker-api/internal/dto"
	mymiddleware "expense-tracker-api/internal/http/middleware"
	"expense-tracker-api/internal/logger"
	"expense-tracker-api/internal/service"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type BudgetHandler struct {
	budgetService *service.BudgetService
}

func NewBudgetHandler(budgetService *service.BudgetService) *BudgetHandler {
	return &BudgetHandler{
		budgetService: budgetService,
	}
}

func (h *BudgetHandler) Create(c echo.Context) error {
	var req dto.CreateBudgetRequest

	if err := c.Bind(&req); err != nil {
		logger.Log.WithError(err).Error("bind create budget request failed")

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
		"user_id":     userID,
		"category_id": req.CategoryID,
		"amount":      req.Amount,
		"month":       req.Month,
	}).Info("create budget request")

	resp, err := h.budgetService.Create(c.Request().Context(), userID, &req)
	if err != nil {
		switch {
		case errors.Is(err, apperror.InvalidRequestBody),
			errors.Is(err, apperror.ErrUserRequired),
			errors.Is(err, apperror.ErrIDRequired),
			errors.Is(err, apperror.ErrInvalidBudgetAmount),
			errors.Is(err, apperror.ErrInvalidMonth):
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
		default:
			logger.Log.WithFields(logrus.Fields{
				"user_id": userID,
				"error":   err.Error(),
			}).Error("create budget failed")

			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": apperror.InternalServer.Error(),
			})
		}
	}

	return c.JSON(http.StatusCreated, resp)
}

func (h *BudgetHandler) List(c echo.Context) error {
	userID, err := mymiddleware.GetUserID(c)
	if err != nil {
		logger.Log.WithError(err).Error("get user id from context failed")

		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": apperror.Unauthorized.Error(),
		})
	}

	resp, err := h.budgetService.ListByUserID(c.Request().Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, apperror.ErrUserRequired):
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
		default:
			logger.Log.WithFields(logrus.Fields{
				"user_id": userID,
				"error":   err.Error(),
			}).Error("list budgets failed")

			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": apperror.InternalServer.Error(),
			})
		}
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *BudgetHandler) Status(c echo.Context) error {
	userID, err := mymiddleware.GetUserID(c)
	if err != nil {
		logger.Log.WithError(err).Error("get user id from context failed")

		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": apperror.Unauthorized.Error(),
		})
	}

	month := strings.TrimSpace(c.QueryParam("month"))

	resp, err := h.budgetService.GetStatus(c.Request().Context(), userID, month)
	if err != nil {
		switch {
		case errors.Is(err, apperror.ErrUserRequired),
			errors.Is(err, apperror.ErrInvalidMonth):
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
		default:
			logger.Log.WithFields(logrus.Fields{
				"user_id": userID,
				"month":   month,
				"error":   err.Error(),
			}).Error("get budget status failed")

			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": apperror.InternalServer.Error(),
			})
		}
	}

	return c.JSON(http.StatusOK, resp)
}
