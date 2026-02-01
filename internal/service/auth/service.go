package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/Dreker052/productivity-app.git/internal/models"
	"github.com/Dreker052/productivity-app.git/internal/repository"
	"github.com/Dreker052/productivity-app.git/internal/utils"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

var (
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidOrExpired   = errors.New("invalid or expired refresh token")
	ErrUserNotVerified    = errors.New("user is not verified")
)

type authService struct {
	jwtSecret   string
	userRepo    repository.UserRepository
	queueClient *asynq.Client
	logger      *slog.Logger
	serverAddr  string
}

func NewAuthService(userRepo repository.UserRepository, logger *slog.Logger, jwtSecret, serverAddr string, queueClient *asynq.Client) *authService {
	return &authService{
		userRepo:    userRepo,
		logger:      logger,
		queueClient: queueClient,
		jwtSecret:   jwtSecret,
		serverAddr:  serverAddr,
	}
}

type EmailPayload struct {
	ToEmail string `json:"to_email"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

const TypeEmailDelivery = "email:delivery"

func (s *authService) Register(ctx context.Context, user *models.User) error {

	existingUser, err := s.userRepo.GetByEmail(ctx, user.Email)
	if err != nil {
		s.logger.Error("[AuthService Register] failed to check existing user", slog.String("error", err.Error()))
		return err
	}
	if existingUser != nil {
		s.logger.Warn("[AuthService Register] user already exists", slog.String("email", user.Email))
		return ErrUserExists
	}

	hashedPass, err := utils.HashPassword(user.Password)
	if err != nil {
		s.logger.Error("[AuthService Register] failed to hash password", slog.String("error", err.Error()))
		return err
	}

	user.ID = strings.ToUpper(uuid.New().String())
	user.Password = hashedPass

	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.Error("[AuthService Register] failed to create user", slog.String("error", err.Error()))
		return err
	}

	verificationToken := uuid.New().String()

	err = s.userRepo.SaveVerificationToken(ctx, verificationToken, user.ID, 1*time.Hour)
	if err != nil {
		s.logger.Error("[AuthService Register] failed to save verification token", slog.String("error", err.Error()))
		return err
	}

	link := fmt.Sprintf("http://%s/api/auth/verify-email?token=%s", s.serverAddr, verificationToken)

	body := fmt.Sprintf(`
		<h2>Добро пожаловать в Productivity App, %s!</h2>
		<p>Пожалуйста, подтвердите вашу почту, нажав на ссылку ниже:</p>
		<p><a href="%s">Подтвердить регистрацию</a></p>
		<br>
		<small>Если вы не регистрировались, просто проигнорируйте это письмо.</small>
	`, user.Name, link)

	payload := EmailPayload{
		ToEmail: user.Email,
		Subject: "Подтверждение почты",
		Body:    body,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		s.logger.Error("[AuthService Register] failed to marshal email payload", slog.String("error", err.Error()))
		return nil
	}

	task := asynq.NewTask(TypeEmailDelivery, payloadBytes)

	info, err := s.queueClient.Enqueue(task, asynq.Queue("emails"))
	if err != nil {
		s.logger.Error("[AuthService Register] failed to enqueue email task", slog.String("error", err.Error()))
	} else {
		s.logger.Info("[AuthService Register] Enqueued verification email", slog.String("task_id", info.ID), slog.String("email", user.Email))
	}

	return nil
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

	if !user.IsVerified {
		return "", "", ErrUserNotVerified
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

func (s *authService) VerifyEmail(ctx context.Context, token string) error {
	return s.userRepo.VerifyEmail(ctx, token)
}
