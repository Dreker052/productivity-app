package user

import (
	"log/slog"

	"github.com/Dreker052/productivity-app.git/internal/service"
)

type userHandler struct {
	service service.UserService
	logger  *slog.Logger
}

func NewUserHandler(service service.UserService, logger *slog.Logger) *userHandler {
	return &userHandler{
		service: service,
		logger:  logger,
	}
}
