package main

import (
	"expense-tracker-api/internal/config"
	"expense-tracker-api/internal/db"
	httppkg "expense-tracker-api/internal/http"
	"expense-tracker-api/internal/http/handler"
	"expense-tracker-api/internal/http/middleware"
	"expense-tracker-api/internal/logger"
	"expense-tracker-api/internal/repository"
	"expense-tracker-api/internal/service"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/labstack/echo/v4"
	"log"
)

func main() {
	logger.Init()

	cfg := config.LoadConfig()

	database, err := db.NewPostgres(cfg)
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}
	defer database.Close()

	runMigrations(cfg)

	e := echo.New()

	userRepo := repository.NewUserRepository(database)
	transactionRepo := repository.NewTransactionRepository(database)
	categoryRepo := repository.NewCategoryRepository(database)
	summaryRepo := repository.NewSummaryRepository(database)
	budgetRepo := repository.NewBudgetRepository(database)

	authService := service.NewAuthService(
		userRepo,
		cfg.JWTAccessSecret,
		cfg.JWTRefreshSecret,
	)
	userService := service.NewUserService(userRepo)
	transactionService := service.NewTransactionService(transactionRepo)
	categoryService := service.NewCategoryService(categoryRepo)
	summaryService := service.NewSummaryService(summaryRepo)
	budgetService := service.NewBudgetService(budgetRepo)
	aiService := service.NewAIService(
		transactionService,
		categoryService,
		summaryService,
		budgetService,
		cfg.OpenAIAPIKey,
		cfg.OpenAIModel,
	)

	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	transactionHandler := handler.NewTransactionHandler(transactionService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	summaryHandler := handler.NewSummaryHandler(summaryService)
	budgetHandler := handler.NewBudgetHandler(budgetService)
	aiHandler := handler.NewAIHandler(aiService)
	jwtMiddleware := middleware.NewJWTMiddleware(authService)

	httppkg.Routes(
		e,
		authHandler,
		userHandler,
		transactionHandler,
		categoryHandler,
		summaryHandler,
		budgetHandler,
		aiHandler,
		jwtMiddleware,
	)

	logger.Log.WithField("port", cfg.AppPort).Info("server started")

	if err := e.Start(":" + cfg.AppPort); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func runMigrations(cfg *config.Config) {
	dsn := "postgres://" + cfg.DBUser + ":" + cfg.DBPassword +
		"@" + cfg.DBHost + ":" + cfg.DBPort + "/" + cfg.DBName + "?sslmode=disable"

	m, err := migrate.New(
		"file://db/migrations",
		dsn,
	)
	if err != nil {
		logger.Log.WithError(err).Fatal("failed to init migrations")
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Log.WithError(err).Fatal("failed to apply migrations")
	}

	logger.Log.Info("migrations applied successfully")
}
