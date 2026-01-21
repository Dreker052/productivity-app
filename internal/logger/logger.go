package logger

import (
	"log/slog"
	"os"

	"github.com/Dreker052/productivity-app.git/internal/config"
)

func SetupLogger(cfg *config.Config) *slog.Logger {
	var logger *slog.Logger

	switch cfg.Env {
	case "local":
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	case "prod":
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	default:
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	}

	return logger

}
