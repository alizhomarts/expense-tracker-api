package handler

import (
	"errors"
	"github.com/google/uuid"
	"net/http"

	"expense-tracker-api/internal/apperror"
	"expense-tracker-api/internal/dto"
	mymiddleware "expense-tracker-api/internal/http/middleware"
	"expense-tracker-api/internal/logger"
	"expense-tracker-api/internal/service"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type TransactionHandler struct {
	service *service.TransactionService
}

func NewTransactionHandler(service *service.TransactionService) *TransactionHandler {
	return &TransactionHandler{
		service: service,
	}
}

// CreateTransaction godoc
// @Summary Create transaction
// @Tags transactions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.CreateTransactionRequest true "Transaction"
// @Success 201 {object} dto.TransactionResponse
// @Failure 400 {object} map[string]string
// @Router /transactions [post]
func (h *TransactionHandler) Create(c echo.Context) error {
	var req dto.CreateTransactionRequest

	if err := c.Bind(&req); err != nil {
		logger.Log.WithError(err).Error("bind create transaction request failed")

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
		"type":        req.Type,
		"amount":      req.Amount,
		"category_id": req.CategoryID,
	}).Info("create transaction request")

	resp, err := h.service.Create(c.Request().Context(), userID, &req)
	if err != nil {
		switch {
		case errors.Is(err, apperror.ErrUserRequired),
			errors.Is(err, apperror.ErrInvalidTransactionType),
			errors.Is(err, apperror.ErrInvalidAmount),
			errors.Is(err, apperror.ErrInvalidDescription),
			errors.Is(err, apperror.InvalidRequestBody):
			logger.Log.WithFields(logrus.Fields{
				"user_id": userID,
				"error":   err.Error(),
			}).Warn("create transaction validation failed")

			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
		default:
			logger.Log.WithFields(logrus.Fields{
				"user_id": userID,
				"error":   err.Error(),
			}).Error("create transaction failed")

			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": apperror.InternalServer.Error(),
			})
		}
	}

	logger.Log.WithFields(logrus.Fields{
		"user_id":        userID,
		"transaction_id": resp.ID,
	}).Info("transaction created successfully")

	return c.JSON(http.StatusCreated, resp)
}

// GetTransaction godoc
// @Summary Get transaction by ID
// @Tags transactions
// @Security BearerAuth
// @Produce json
// @Param id path string true "Transaction ID"
// @Success 200 {object} dto.TransactionResponse
// @Failure 404 {object} map[string]string
// @Router /transactions/{id} [get]
func (h *TransactionHandler) GetByID(c echo.Context) error {
	userID, err := mymiddleware.GetUserID(c)
	if err != nil {
		logger.Log.WithError(err).Error("get user id from context failed")

		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": apperror.Unauthorized.Error(),
		})
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"user_id":   userID,
			"param_id":  c.Param("id"),
			"parse_err": err.Error(),
		}).Warn("invalid transaction id")

		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": apperror.ErrIDRequired.Error(),
		})
	}

	logger.Log.WithFields(logrus.Fields{
		"user_id":        userID,
		"transaction_id": id,
	}).Info("get transaction by id request")

	resp, err := h.service.GetByID(c.Request().Context(), id, userID)
	if err != nil {
		switch {
		case errors.Is(err, apperror.ErrIDRequired),
			errors.Is(err, apperror.ErrUserRequired):
			logger.Log.WithFields(logrus.Fields{
				"user_id":        userID,
				"transaction_id": id,
				"error":          err.Error(),
			}).Warn("get transaction by id validation failed")

			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
		case errors.Is(err, apperror.ErrTransactionNotFound):
			logger.Log.WithFields(logrus.Fields{
				"user_id":        userID,
				"transaction_id": id,
			}).Warn("transaction not found")

			return c.JSON(http.StatusNotFound, map[string]string{
				"error": err.Error(),
			})
		default:
			logger.Log.WithFields(logrus.Fields{
				"user_id":        userID,
				"transaction_id": id,
				"error":          err.Error(),
			}).Error("get transaction by id failed")

			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": apperror.InternalServer.Error(),
			})
		}
	}

	logger.Log.WithFields(logrus.Fields{
		"user_id":        userID,
		"transaction_id": id,
	}).Info("transaction returned successfully")

	return c.JSON(http.StatusOK, resp)
}

// ListTransactions godoc
// @Summary Get transactions
// @Tags transactions
// @Security BearerAuth
// @Produce json
// @Success 200 {array} dto.TransactionResponse
// @Router /transactions [get]
func (h *TransactionHandler) List(c echo.Context) error {
	userID, err := mymiddleware.GetUserID(c)
	if err != nil {
		logger.Log.WithError(err).Error("get user id from context failed")

		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": apperror.Unauthorized.Error(),
		})
	}

	logger.Log.WithFields(logrus.Fields{
		"user_id": userID,
	}).Info("list transactions request")

	resp, err := h.service.ListByUserID(c.Request().Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, apperror.ErrUserRequired):
			logger.Log.WithFields(logrus.Fields{
				"user_id": userID,
				"error":   err.Error(),
			}).Warn("list transactions validation failed")

			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
		default:
			logger.Log.WithFields(logrus.Fields{
				"user_id": userID,
				"error":   err.Error(),
			}).Error("list transactions failed")

			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": apperror.InternalServer.Error(),
			})
		}
	}

	logger.Log.WithFields(logrus.Fields{
		"user_id": userID,
		"count":   len(resp),
	}).Info("transactions listed successfully")

	return c.JSON(http.StatusOK, resp)
}

// UpdateTransaction godoc
// @Summary Update transaction
// @Tags transactions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Transaction ID"
// @Param request body dto.UpdateTransactionRequest true "Update data"
// @Success 200 {object} dto.TransactionResponse
// @Router /transactions/{id} [put]
func (h *TransactionHandler) Update(c echo.Context) error {
	var req dto.UpdateTransactionRequest

	if err := c.Bind(&req); err != nil {
		logger.Log.WithError(err).Error("bind update transaction request failed")

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

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"user_id":   userID,
			"param_id":  c.Param("id"),
			"parse_err": err.Error(),
		}).Warn("invalid transaction id")

		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": apperror.ErrIDRequired.Error(),
		})
	}

	logger.Log.WithFields(logrus.Fields{
		"user_id":        userID,
		"transaction_id": id,
		"type":           req.Type,
		"amount":         req.Amount,
		"category_id":    req.CategoryID,
	}).Info("update transaction request")

	resp, err := h.service.Update(c.Request().Context(), id, userID, &req)
	if err != nil {
		switch {
		case errors.Is(err, apperror.ErrIDRequired),
			errors.Is(err, apperror.ErrUserRequired),
			errors.Is(err, apperror.ErrInvalidTransactionType),
			errors.Is(err, apperror.ErrInvalidAmount),
			errors.Is(err, apperror.ErrInvalidDescription),
			errors.Is(err, apperror.InvalidRequestBody):
			logger.Log.WithFields(logrus.Fields{
				"user_id":        userID,
				"transaction_id": id,
				"error":          err.Error(),
			}).Warn("update transaction validation failed")

			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
		case errors.Is(err, apperror.ErrTransactionNotFound):
			logger.Log.WithFields(logrus.Fields{
				"user_id":        userID,
				"transaction_id": id,
			}).Warn("update transaction not found")

			return c.JSON(http.StatusNotFound, map[string]string{
				"error": err.Error(),
			})
		default:
			logger.Log.WithFields(logrus.Fields{
				"user_id":        userID,
				"transaction_id": id,
				"error":          err.Error(),
			}).Error("update transaction failed")

			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": apperror.InternalServer.Error(),
			})
		}
	}

	logger.Log.WithFields(logrus.Fields{
		"user_id":        userID,
		"transaction_id": id,
	}).Info("transaction updated successfully")

	return c.JSON(http.StatusOK, resp)
}

// DeleteTransaction godoc
// @Summary Delete transaction
// @Tags transactions
// @Security BearerAuth
// @Param id path string true "Transaction ID"
// @Success 204
// @Router /transactions/{id} [delete]
func (h *TransactionHandler) Delete(c echo.Context) error {
	userID, err := mymiddleware.GetUserID(c)
	if err != nil {
		logger.Log.WithError(err).Error("get user id from context failed")

		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": apperror.Unauthorized.Error(),
		})
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"user_id":   userID,
			"param_id":  c.Param("id"),
			"parse_err": err.Error(),
		}).Warn("invalid transaction id")

		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": apperror.ErrIDRequired.Error(),
		})
	}

	logger.Log.WithFields(logrus.Fields{
		"user_id":        userID,
		"transaction_id": id,
	}).Info("delete transaction request")

	err = h.service.Delete(c.Request().Context(), id, userID)
	if err != nil {
		switch {
		case errors.Is(err, apperror.ErrIDRequired),
			errors.Is(err, apperror.ErrUserRequired):
			logger.Log.WithFields(logrus.Fields{
				"user_id":        userID,
				"transaction_id": id,
				"error":          err.Error(),
			}).Warn("delete transaction validation failed")

			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
		case errors.Is(err, apperror.ErrTransactionNotFound):
			logger.Log.WithFields(logrus.Fields{
				"user_id":        userID,
				"transaction_id": id,
			}).Warn("delete transaction not found")

			return c.JSON(http.StatusNotFound, map[string]string{
				"error": err.Error(),
			})
		default:
			logger.Log.WithFields(logrus.Fields{
				"user_id":        userID,
				"transaction_id": id,
				"error":          err.Error(),
			}).Error("delete transaction failed")

			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": apperror.InternalServer.Error(),
			})
		}
	}

	logger.Log.WithFields(logrus.Fields{
		"user_id":        userID,
		"transaction_id": id,
	}).Info("transaction deleted successfully")

	return c.NoContent(http.StatusNoContent)
}
