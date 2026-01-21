package auth

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/Dreker052/productivity-app.git/internal/models"
	"github.com/Dreker052/productivity-app.git/internal/repository"
	"github.com/Dreker052/productivity-app.git/internal/utils"
	"github.com/google/uuid"
)

var (
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidOrExpired   = errors.New("invalid or expired refresh token")
)

type authService struct {
	jwtSecret string
	userRepo  repository.UserRepository
	logger    *slog.Logger
}

func NewAuthService(userRepo repository.UserRepository, logger *slog.Logger, jwtSecret string) *authService {
	return &authService{
		userRepo:  userRepo,
		logger:    logger,
		jwtSecret: jwtSecret,
	}
}

func (s *authService) Register(ctx context.Context, user *models.User) (string, string, error) {

	existingUser, err := s.userRepo.GetByEmail(ctx, user.Email)
	if err != nil {
		s.logger.Error("[AuthService Register] failed to check existing user", slog.String("error", err.Error()))
		return "", "", err
	}
	if existingUser != nil {
		s.logger.Warn("[AuthService Register] user already exists", slog.String("email", user.Email))
		return "", "", ErrUserExists
	}

	hashedPass, err := utils.HashPassword(user.Password)
	if err != nil {
		s.logger.Error("[AuthService Register] failed to hash password", slog.String("error", err.Error()))
		return "", "", err
	}

	user.ID = strings.ToUpper(uuid.New().String())
	user.Password = hashedPass

	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.Error("[AuthService Register] failed to create user", slog.String("error", err.Error()))
		return "", "", err
	}

	return utils.GenerateTokens(user.ID, s.jwtSecret)
}

func (s *authService) Login(ctx context.Context, email, password string) (string, string, error) {

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		s.logger.Error("[AuthService Login] failed to get user by email", slog.String("error", err.Error()))
		return "", "", err
	}

	if user == nil {
		s.logger.Warn("[AuthService Login] user not found", slog.String("email", email))
		return "", "", ErrInvalidCredentials
	}

	if !utils.CheckPasswordHash(password, user.Password) {
		s.logger.Warn("[AuthService Login] invalid credentials", slog.String("email", email))
		return "", "", ErrInvalidCredentials
	}

	return utils.GenerateTokens(user.ID, s.jwtSecret)
}

func (s *authService) RefreshTokens(ctx context.Context, refreshToken string) (string, string, error) {

	userID, err := utils.ValidateToken(refreshToken, s.jwtSecret)
	if err != nil {
		return "", "", ErrInvalidOrExpired
	}

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil || user == nil {
		return "", "", ErrUserNotFound
	}

	return utils.GenerateTokens(userID, s.jwtSecret)
}
