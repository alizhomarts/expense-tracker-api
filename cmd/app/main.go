package main

import (
	"expense-tracker-api/internal/config"
	"expense-tracker-api/internal/db"
	httppkg "expense-tracker-api/internal/http"
	"expense-tracker-api/internal/http/handler"
	"expense-tracker-api/internal/http/middleware"
	"expense-tracker-api/internal/repository"
	"expense-tracker-api/internal/service"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/labstack/echo/v4"
	"log"
)

func main() {
	cfg := config.LoadConfig()

	database, err := db.NewPostgres(cfg)
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}
	defer database.Close()
	runMigrations(cfg)

	e := echo.New()

	userRepo := repository.NewUserRepository(database)

	authService := service.NewAuthService(
		userRepo,
		cfg.JWTAccessSecret,
		cfg.JWTRefreshSecret,
	)

	authHandler := handler.NewAuthHandler(authService)
	jwtMiddleware := middleware.NewJWTMiddleware(authService)

	httppkg.Routes(e, authHandler, jwtMiddleware)

	log.Printf("server started on :%s", cfg.AppPort)

	if err := e.Start(":" + cfg.AppPort); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func runMigrations(cfg *config.Config) {
	log.Println("running migrations...")

	dsn := "postgres://" + cfg.DBUser + ":" + cfg.DBPassword +
		"@" + cfg.DBHost + ":" + cfg.DBPort + "/" + cfg.DBName + "?sslmode=disable"

	m, err := migrate.New(
		"file://db/migrations",
		dsn,
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}

	log.Println("migrations applied")
}
