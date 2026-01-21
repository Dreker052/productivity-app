package database

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/Dreker052/productivity-app.git/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func InitDB(cfg *config.Config, logger *slog.Logger) *pgxpool.Pool {

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBSSLMode,
	)

	pgCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		logger.Error("failed to parse pgx config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), pgCfg)
	if err != nil {
		logger.Error("failed to create pgx pool", slog.String("error", err.Error()))
		os.Exit(1)
	}

	err = pool.Ping(context.Background())
	if err != nil {
		logger.Error("failed to ping db", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("connected to db successfully")

	return pool

}
