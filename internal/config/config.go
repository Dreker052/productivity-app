package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	ServerPort string

	Env string

	JWTSecret string
}

func LoadConfig() (*Config, error) {

	if err := godotenv.Load(); err != nil {
		slog.Error("Error loading .env file", slog.String("error", err.Error()))
		return nil, err
	}

	cfg := &Config{
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		DBSSLMode:  os.Getenv("DB_SSLMODE"),

		ServerPort: os.Getenv("SERVER_PORT"),

		Env: os.Getenv("ENV"),

		JWTSecret: os.Getenv("JWT_SECRET"),
	}

	return cfg, nil
}
