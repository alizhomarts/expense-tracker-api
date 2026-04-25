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

type SummaryHandler struct {
	summaryService *service.SummaryService
}

func NewSummaryHandler(summaryService *service.SummaryService) *SummaryHandler {
	return &SummaryHandler{
		summaryService: summaryService,
	}
}

// GetSummary godoc
// @Summary Get summary
// @Tags summary
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.SummaryResponse
// @Router /summary [get]
func (h *SummaryHandler) GetSummary(c echo.Context) error {
	userID, err := mymiddleware.GetUserID(c)
	if err != nil {
		logger.Log.WithError(err).Error("get user id from context failed")

		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": apperror.Unauthorized.Error(),
		})
	}

	logger.Log.WithFields(logrus.Fields{
		"user_id": userID,
	}).Info("get summary request")

	resp, err := h.summaryService.GetSummary(c.Request().Context(), userID)
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
			}).Error("get summary failed")

			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": apperror.InternalServer.Error(),
			})
		}
	}

	return c.JSON(http.StatusOK, resp)
}

// GetCategorySummary godoc
// @Summary Get category summary
// @Tags summary
// @Security BearerAuth
// @Produce json
// @Success 200 {array} dto.CategorySummaryResponse
// @Router /summary/categories [get]
func (h *SummaryHandler) GetCategorySummary(c echo.Context) error {
	userID, err := mymiddleware.GetUserID(c)
	if err != nil {
		logger.Log.WithError(err).Error("get user id from context failed")

		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": apperror.Unauthorized.Error(),
		})
	}

	logger.Log.WithFields(logrus.Fields{
		"user_id": userID,
	}).Info("get category summary request")

	resp, err := h.summaryService.GetCategorySummary(c.Request().Context(), userID)
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
			}).Error("get category summary failed")

			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": apperror.InternalServer.Error(),
			})
		}
	}

	return c.JSON(http.StatusOK, resp)
}

// GetMonthlySummary godoc
// @Summary Get monthly summary
// @Tags summary
// @Security BearerAuth
// @Produce json
// @Param month query string true "Month YYYY-MM"
// @Success 200 {object} dto.MonthlySummaryResponse
// @Router /summary/monthly [get]
func (h *SummaryHandler) GetMonthlySummary(c echo.Context) error {
	userID, err := mymiddleware.GetUserID(c)
	if err != nil {
		logger.Log.WithError(err).Error("get user id from context failed")

		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": apperror.Unauthorized.Error(),
		})
	}

	month := c.QueryParam("month")

	logger.Log.WithFields(logrus.Fields{
		"user_id": userID,
		"month":   month,
	}).Info("get monthly summary request")

	resp, err := h.summaryService.GetMonthlySummary(c.Request().Context(), userID, month)
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
			}).Error("get monthly summary failed")

			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": apperror.InternalServer.Error(),
			})
		}
	}

	return c.JSON(http.StatusOK, resp)
}
