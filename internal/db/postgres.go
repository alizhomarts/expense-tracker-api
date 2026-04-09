package db

import (
	"context"
	"expense-tracker-api/internal/config"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
)

func NewPostgres(cfg *config.Config) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect db: %v", err)
	}

	if err = pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping db: %v", err)
	}

	log.Println("connected to postgres")
	return pool, nil
}
