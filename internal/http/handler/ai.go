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

type AIHandler struct {
	aiService *service.AIService
}

func NewAIHandler(aiService *service.AIService) *AIHandler {
	return &AIHandler{
		aiService: aiService,
	}
}

// Parse godoc
// @Summary Parse transaction from text
// @Tags ai
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.AIParseRequest true "Text"
// @Success 200 {object} dto.AIParseResponse
// @Router /ai/parse [post]
func (h *AIHandler) Parse(c echo.Context) error {
	var req dto.AIParseRequest

	if err := c.Bind(&req); err != nil {
		logger.Log.WithError(err).Error("bind ai parse request failed")

		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": apperror.InvalidRequestBody.Error(),
		})
	}

	req.Text = strings.TrimSpace(req.Text)
	if req.Text == "" {
		logger.Log.Warn("ai parse request text is empty")

		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": apperror.InvalidRequestBody.Error(),
		})
	}

	logger.Log.WithFields(logrus.Fields{
		"text": req.Text,
	}).Info("ai parse request received")

	resp, err := h.aiService.ParseTransactionText(c.Request().Context(), req.Text)
	if err != nil {
		logger.Log.WithError(err).Error("ai parse failed")

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": apperror.InternalServer.Error(),
		})
	}

	logger.Log.WithFields(logrus.Fields{
		"intent":   resp.Intent,
		"amount":   resp.Amount,
		"category": resp.Category,
	}).Info("ai parse completed successfully")

	return c.JSON(http.StatusOK, resp)
}

func (h *AIHandler) ParseAndCreate(c echo.Context) error {
	var req dto.AIParseRequest

	if err := c.Bind(&req); err != nil {
		logger.Log.WithError(err).Error("bind ai parse-and-create request failed")

		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": apperror.InvalidRequestBody.Error(),
		})
	}

	req.Text = strings.TrimSpace(req.Text)
	if req.Text == "" {
		logger.Log.Warn("ai parse-and-create request text is empty")

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
		"text":    req.Text,
	}).Info("ai parse-and-create request received")

	resp, err := h.aiService.ParseAndCreate(c.Request().Context(), userID, req.Text)
	if err != nil {
		switch {
		case errors.Is(err, apperror.ErrUserRequired),
			errors.Is(err, apperror.InvalidRequestBody),
			errors.Is(err, apperror.ErrInvalidTransactionType),
			errors.Is(err, apperror.ErrInvalidAmount),
			errors.Is(err, apperror.ErrInvalidDescription):
			logger.Log.WithFields(logrus.Fields{
				"user_id": userID,
				"error":   err.Error(),
			}).Warn("ai parse-and-create validation failed")

			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
		default:
			logger.Log.WithFields(logrus.Fields{
				"user_id": userID,
				"error":   err.Error(),
			}).Error("ai parse-and-create failed")

			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": apperror.InternalServer.Error(),
			})
		}
	}

	logger.Log.WithFields(logrus.Fields{
		"user_id":        userID,
		"transaction_id": resp.ID,
		"amount":         resp.Amount,
		"type":           resp.Type,
	}).Info("ai parse-and-create completed successfully")

	return c.JSON(http.StatusCreated, resp)
}

// Insights godoc
// @Summary Get AI insights
// @Tags ai
// @Security BearerAuth
// @Produce json
// @Param month query string true "Month YYYY-MM"
// @Success 200 {object} dto.AIInsightsResponse
// @Router /ai/insights [get]
func (h *AIHandler) Insights(c echo.Context) error {
	userID, err := mymiddleware.GetUserID(c)
	if err != nil {
		logger.Log.WithError(err).Error("get user id from context failed")

		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": apperror.Unauthorized.Error(),
		})
	}

	month := strings.TrimSpace(c.QueryParam("month"))

	logger.Log.WithFields(logrus.Fields{
		"user_id": userID,
		"month":   month,
	}).Info("ai insights request received")

	resp, err := h.aiService.GenerateInsights(c.Request().Context(), userID, month)
	if err != nil {
		switch {
		case errors.Is(err, apperror.ErrUserRequired),
			errors.Is(err, apperror.ErrInvalidMonth):
			logger.Log.WithFields(logrus.Fields{
				"user_id": userID,
				"month":   month,
				"error":   err.Error(),
			}).Warn("ai insights validation failed")

			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
		default:
			logger.Log.WithFields(logrus.Fields{
				"user_id": userID,
				"month":   month,
				"error":   err.Error(),
			}).Error("ai insights failed")

			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": apperror.InternalServer.Error(),
			})
		}
	}

	return c.JSON(http.StatusOK, resp)
}

// BudgetAlerts godoc
// @Summary Get AI budget alerts
// @Tags ai
// @Security BearerAuth
// @Produce json
// @Param month query string true "Month YYYY-MM"
// @Success 200 {object} dto.AIBudgetAlertsResponse
// @Router /ai/budget-alerts [get]
func (h *AIHandler) BudgetAlerts(c echo.Context) error {
	userID, err := mymiddleware.GetUserID(c)
	if err != nil {
		logger.Log.WithError(err).Error("get user id from context failed")

		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": apperror.Unauthorized.Error(),
		})
	}

	month := strings.TrimSpace(c.QueryParam("month"))

	logger.Log.WithFields(logrus.Fields{
		"user_id": userID,
		"month":   month,
	}).Info("ai budget alerts request received")

	resp, err := h.aiService.GenerateBudgetAlerts(c.Request().Context(), userID, month)
	if err != nil {
		switch {
		case errors.Is(err, apperror.ErrUserRequired),
			errors.Is(err, apperror.ErrInvalidMonth):
			logger.Log.WithFields(logrus.Fields{
				"user_id": userID,
				"month":   month,
				"error":   err.Error(),
			}).Warn("ai budget alerts validation failed")

			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
		default:
			logger.Log.WithFields(logrus.Fields{
				"user_id": userID,
				"month":   month,
				"error":   err.Error(),
			}).Error("ai budget alerts failed")

			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": apperror.InternalServer.Error(),
			})
		}
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *AIHandler) ParseReceipt(c echo.Context) error {
	var req dto.AIReceiptParseRequest

	if err := c.Bind(&req); err != nil {
		logger.Log.WithError(err).Error("bind ai receipt parse request failed")

		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": apperror.InvalidRequestBody.Error(),
		})
	}

	req.Text = strings.TrimSpace(req.Text)
	if req.Text == "" {
		logger.Log.Warn("ai receipt parse request text is empty")

		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": apperror.InvalidRequestBody.Error(),
		})
	}

	logger.Log.WithFields(logrus.Fields{
		"text": req.Text,
	}).Info("ai receipt parse request received")

	resp, err := h.aiService.ParseReceiptText(c.Request().Context(), req.Text)
	if err != nil {
		logger.Log.WithError(err).Error("ai receipt parse failed")

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": apperror.InternalServer.Error(),
		})
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *AIHandler) ReceiptToTransaction(c echo.Context) error {
	var req dto.AIReceiptParseRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": apperror.InvalidRequestBody.Error(),
		})
	}

	userID, err := mymiddleware.GetUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "unauthorized",
		})
	}

	resp, err := h.aiService.ReceiptToTransaction(c.Request().Context(), userID, req.Text)
	if err != nil {
		logger.Log.WithError(err).Error("receipt to transaction failed")

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": apperror.InternalServer.Error(),
		})
	}

	return c.JSON(http.StatusCreated, resp)
}

// ReceiptToTransactions godoc
// @Summary Convert receipt to multiple transactions
// @Tags ai
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.AIReceiptParseRequest true "Receipt text"
// @Success 201 {array} dto.TransactionResponse
// @Router /ai/receipt-to-transactions [post]
func (h *AIHandler) ReceiptToTransactions(c echo.Context) error {
	var req dto.AIReceiptParseRequest

	if err := c.Bind(&req); err != nil {
		logger.Log.WithError(err).Error("bind receipt-to-transactions request failed")

		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": apperror.InvalidRequestBody.Error(),
		})
	}

	req.Text = strings.TrimSpace(req.Text)
	if req.Text == "" {
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
	}).Info("receipt-to-transactions request received")

	resp, err := h.aiService.ReceiptToTransactions(
		c.Request().Context(),
		userID,
		req.Text,
	)
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"user_id": userID,
			"error":   err.Error(),
		}).Error("receipt-to-transactions failed")

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": apperror.InternalServer.Error(),
		})
	}

	return c.JSON(http.StatusCreated, resp)
}
