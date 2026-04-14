package http

import (
	"expense-tracker-api/internal/http/handler"
	"expense-tracker-api/internal/http/middleware"
	"github.com/labstack/echo/v4"
	"net/http"
)

func Routes(
	e *echo.Echo,
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	transactionHandler *handler.TransactionHandler,
	categoryHandler *handler.CategoryHandler,
	summaryHandler *handler.SummaryHandler,
	budgetHandler *handler.BudgetHandler,
	aiHandler *handler.AIHandler,
	jwtMiddleware *middleware.JWTMiddleware,
) {
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "ok",
		})
	})

	api := e.Group("/api/v1")

	auth := api.Group("/auth")
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)
	auth.POST("/refresh", authHandler.Refresh)

	protected := api.Group("")
	protected.Use(jwtMiddleware.Handle)

	protected.GET("/me", func(c echo.Context) error {
		userID := c.Get(string(middleware.ContextUserIDKey))

		return c.JSON(http.StatusOK, map[string]any{
			"message": "authorized",
			"user_id": userID,
		})
	})

	users := protected.Group("/users")
	users.GET("/me", userHandler.GetMe)

	transactions := protected.Group("/transactions")
	transactions.POST("", transactionHandler.Create)
	transactions.GET("", transactionHandler.List)
	transactions.GET("/:id", transactionHandler.GetByID)
	transactions.PUT("/:id", transactionHandler.Update)
	transactions.DELETE("/:id", transactionHandler.Delete)

	categories := protected.Group("/categories")
	categories.POST("", categoryHandler.Create)
	categories.GET("", categoryHandler.List)

	summary := protected.Group("/summary")
	summary.GET("", summaryHandler.GetSummary)
	summary.GET("/categories", summaryHandler.GetCategorySummary)
	summary.GET("/monthly", summaryHandler.GetMonthlySummary)

	budgets := protected.Group("/budgets")
	budgets.POST("", budgetHandler.Create)
	budgets.GET("", budgetHandler.List)
	budgets.GET("/status", budgetHandler.Status)

	ai := protected.Group("/ai")
	ai.POST("/parse", aiHandler.Parse)
	ai.POST("/parse-and-create", aiHandler.ParseAndCreate)
	ai.GET("/insights", aiHandler.Insights)
	ai.GET("/budget-alerts", aiHandler.BudgetAlerts)
	ai.POST("/parse-receipt", aiHandler.ParseReceipt)
	ai.POST("/receipt-to-transaction", aiHandler.ReceiptToTransaction)
}
