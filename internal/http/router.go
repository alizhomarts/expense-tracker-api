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

	me := api.Group("")
	me.Use(jwtMiddleware.Handle)

	me.GET("/me", func(c echo.Context) error {
		userID := c.Get(middleware.ContextUserIDKey)

		return c.JSON(http.StatusOK, map[string]string{
			"message": "authorized",
			"user_id": userID.(string),
		})
	})
}
