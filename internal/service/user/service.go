package user

import (
	"context"
	"errors"
	"log/slog"

	"github.com/Dreker052/productivity-app/internal/models"
	"github.com/Dreker052/productivity-app/internal/repository"
	"github.com/Dreker052/productivity-app/internal/utils"
)

type userService struct {
	userRepo repository.UserRepository
	logger   *slog.Logger
}

func NewUserService(userRepo repository.UserRepository, logger *slog.Logger) *userService {
	return &userService{
		userRepo: userRepo,
		logger:   logger,
	}
}

func (s *userService) Create(ctx context.Context, user *models.User) error {
	return s.userRepo.Create(ctx, user)
}

func (s *userService) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return s.userRepo.GetByEmail(ctx, email)
}

func (s *userService) GetByID(ctx context.Context, id string) (*models.User, error) {
	return s.userRepo.GetUserByID(ctx, id)
}

func (s *userService) ChangeName(ctx context.Context, userID, newName string) error {
	return s.userRepo.ChangeName(ctx, userID, newName)
}

func (s *userService) ChangeEmail(ctx context.Context, userID, newEmail string) error {
	exist, err := s.userRepo.GetByEmail(ctx, newEmail)
	if err != nil {
		s.logger.Error("[UserService ChangeEmail] failed to get user by email", slog.String("error", err.Error()))
	}
	if exist != nil && userID != exist.ID {
		s.logger.Warn("[UserService ChangeEmail] failed to change email. This email is already in use", slog.String("userID", userID))
		return errors.New("email already taken")

	}
	return s.userRepo.ChangeEmail(ctx, userID, newEmail)
}

func (s *userService) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		s.logger.Error("[UserService ChangePassword] failed to get user by id", slog.String("error", err.Error()))
		return err
	}
	if user == nil {
		s.logger.Warn("[UserService ChangePassword] failed to change password, user not found")
		return errors.New("user not found")
	}

	if !utils.CheckPasswordHash(oldPassword, user.Password) {
		s.logger.Warn("[UserService ChangePassword] failed to change password, old password is invalid")
		return errors.New("invalid old password")
	}

	newHash, err := utils.HashPassword(newPassword)
	if err != nil {
		s.logger.Error("[UserService ChangePassword] failed to hash new password", slog.String("error", err.Error()))
		return err
	}

	return s.userRepo.ChangePassword(ctx, userID, newHash)
}
