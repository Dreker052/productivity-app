package user

import (
	"context"
	"log/slog"

	"github.com/Dreker052/productivity-app.git/internal/models"
	"github.com/Dreker052/productivity-app.git/internal/repository"
)

type userService struct {
	repo   repository.UserRepository
	logger *slog.Logger
}

func NewUserService(repo repository.UserRepository, logger *slog.Logger) *userService {
	return &userService{
		repo:   repo,
		logger: logger,
	}
}

func (s *userService) Create(ctx context.Context, user *models.User) error {
	return s.repo.Create(ctx, user)
}

func (s *userService) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return s.repo.GetByEmail(ctx, email)
}
