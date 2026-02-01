package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/Dreker052/productivity-app.git/internal/models"
	"github.com/Dreker052/productivity-app.git/internal/service"
	"github.com/gin-gonic/gin"
)

type AuthInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type AuthResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type RegisterInput struct {
	Email    string `json:"email" binding:"required,email"`
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

type RefreshInput struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type authHandler struct {
	AuthService service.AuthService
	logger      *slog.Logger
}

func NewAuthHandler(authService service.AuthService, logger *slog.Logger) *authHandler {
	return &authHandler{
		AuthService: authService,
		logger:      logger,
	}
}

func (h *authHandler) Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		h.logger.Error("failed to bind json", slog.String("error", err.Error()))
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	user := &models.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: input.Password,
	}

	if err := h.AuthService.Register(ctx, user); err != nil {
		if errors.Is(err, service.ErrUserAlreadyExists) {
			c.JSON(409, gin.H{"error": "User already exists"})
			return
		}
		h.logger.Error("failed to register user", slog.String("error", err.Error()))
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	h.logger.Info("user registered successfully", slog.String("email", input.Email))

	c.JSON(http.StatusOK, gin.H{
		"message": "Registration successful. Please check your email to verify account.",
	})
}

func (h *authHandler) Login(c *gin.Context) {
	var input AuthInput

	if err := c.ShouldBindJSON(&input); err != nil {
		h.logger.Error("Failed to bind JSON", slog.String("error", err.Error()))
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	accessToken, refreshToken, err := h.AuthService.Login(ctx, input.Email, input.Password)

	if err != nil {

		h.logger.Warn("Failed login attempt",
			slog.String("email", input.Email),
			slog.String("error", err.Error()),
		)

		c.JSON(401, gin.H{"error": "Invalid email or password"})
		return
	}

	h.logger.Info("User logged in successfully", slog.String("email", input.Email))

	c.JSON(http.StatusOK, AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (h *authHandler) RefreshTokens(c *gin.Context) {

	var input RefreshInput

	if err := c.ShouldBindJSON(&input); err != nil {
		h.logger.Error("[AuthHandler RefreshTokens] Failed to bind JSON", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	accessToken, refreshToken, err := h.AuthService.RefreshTokens(ctx, input.RefreshToken)

	if err != nil {
		h.logger.Error("[AuthHandler RefreshTokens] Failed to refresh tokens, invalid refresh token", slog.String("error", err.Error()))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

// GET api/auth/verify-email?token=...
func (h *authHandler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		h.logger.Error("[AuthHandler VerifyEmail] Missing verification token")
		c.Data(400, "text/html; charset=utf-8", []byte("<h1>Ошибка: Код не найден</h1>"))
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if err := h.AuthService.VerifyEmail(ctx, token); err != nil {
		h.logger.Error("[AuthHandler VerifyEmail] Failed to verify email", slog.String("error", err.Error()))
		c.Data(400, "text/html; charset=utf-8", []byte(`
			<div style="text-align: center; margin-top: 50px; font-family: sans-serif;">
				<h1 style="color: red;">Ошибка подтверждения</h1>
				<p>Ссылка недействительна или срок её действия истек.</p>
			</div>
		`))
		return
	}

	c.Data(200, "text/html; charset=utf-8", []byte(`
		<div style="text-align: center; margin-top: 50px; font-family: sans-serif;">
			<h1 style="color: green;">Email успешно подтвержден! ✅</h1>
			<p>Теперь вы можете вернуться в приложение Productivity App и войти в свой аккаунт.</p>
		</div>
	`))
}
