package user

import (
	"context"
	"log/slog"
	"time"

	"github.com/Dreker052/productivity-app.git/internal/service"
	"github.com/gin-gonic/gin"
)

type userOutput struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type userHandler struct {
	userService service.UserService
	logger      *slog.Logger
}

func NewUserHandler(userService service.UserService, logger *slog.Logger) *userHandler {
	return &userHandler{
		userService: userService,
		logger:      logger,
	}
}

func (h *userHandler) GetUserData(c *gin.Context) {
	userID := c.MustGet("userID").(string)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	user, err := h.userService.GetByID(ctx, userID)
	if err != nil {
		h.logger.Error("[UserHandler GetByID] Failed to get user data", slog.String("error", err.Error()))
		c.JSON(500, gin.H{"error": "failed to get user data"})
		return
	}

	userOutput := &userOutput{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}

	h.logger.Info("[UserHandler GetByID] User data retrieved successfully", slog.String("userID", user.ID))

	c.JSON(200, userOutput)
}

type ChangeNameInput struct {
	Name string `json:"name" binding:"required"`
}

func (h *userHandler) ChangeName(c *gin.Context) {
	userID := c.MustGet("userID").(string)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var input ChangeNameInput
	if err := c.ShouldBindJSON(&input); err != nil {
		h.logger.Error("[UserHanler ChangeName] failed to parse json", slog.String("error", err.Error()))
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.ChangeName(ctx, userID, input.Name); err != nil {
		h.logger.Error("[UserHandler ChangeName] failed to change name", slog.String("error", err.Error()))
		c.JSON(400, gin.H{"error": "Failed to change name"})
		return
	}

	h.logger.Info("[UserHanler ChangeName] name successfully changed", slog.String("userID", userID))

	c.JSON(200, gin.H{"message": "name successfully changed"})
}

type ChangeEmailInput struct {
	Email string `json:"email" binding:"email,required"`
}

func (h *userHandler) ChangeEmail(c *gin.Context) {
	userID := c.MustGet("userID").(string)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var input ChangeEmailInput
	if err := c.ShouldBindJSON(&input); err != nil {
		h.logger.Error("[UserHanler ChangeEmail] failed to parse json", slog.String("error", err.Error()))
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.ChangeEmail(ctx, userID, input.Email); err != nil {
		h.logger.Error("[UserHandler ChangeEmail] failed to change email", slog.String("error", err.Error()))
		c.JSON(500, gin.H{"error": "Failed to change email"})
		return
	}

	h.logger.Info("[UserHanler ChangeEmail] email successfully changed", slog.String("userID", userID))

	c.JSON(200, gin.H{"message": "email successfully changed"})
}

type ChangePasswordInput struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=8"`
}

func (h *userHandler) ChangePassword(c *gin.Context) {
	userID := c.MustGet("userID").(string)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var input ChangePasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		h.logger.Error("[UserHanler ChangePassword] failed to parse json", slog.String("error", err.Error()))
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.ChangePassword(ctx, userID, input.OldPassword, input.NewPassword); err != nil {
		h.logger.Error("[UserHandler ChangePassword] failed to change password", slog.String("error", err.Error()))
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("[UserHanler ChangePassword] password successfully changed", slog.String("userID", userID))

	c.JSON(200, gin.H{"message": "password successfully changed"})

}
