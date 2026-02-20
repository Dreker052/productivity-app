package yearlygoal

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/Dreker052/productivity-app/internal/models"
	"github.com/Dreker052/productivity-app/internal/service"
	"github.com/gin-gonic/gin"
)

type yearlyGoalHandler struct {
	yearlyGoalService service.YearlyGoalService
	logger            *slog.Logger
}

func NewYearlyGoalHandler(yearlyGoalService service.YearlyGoalService, logger *slog.Logger) *yearlyGoalHandler {
	return &yearlyGoalHandler{
		yearlyGoalService: yearlyGoalService,
		logger:            logger,
	}
}

func (h *yearlyGoalHandler) GetGroups(c *gin.Context) {
	userID := c.MustGet("userID").(string)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	groups, err := h.yearlyGoalService.GetGroups(ctx, userID)
	if err != nil {
		h.logger.Error("Failed to fetch goals", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch goals"})
		return
	}
	c.JSON(http.StatusOK, groups)
}

// POST /api/groups
func (h *yearlyGoalHandler) CreateGroup(c *gin.Context) {
	userID := c.MustGet("userID").(string)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var group models.GoalGroup
	if err := c.ShouldBindJSON(&group); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Creating group", slog.Any("group", group))

	if err := h.yearlyGoalService.CreateGroup(ctx, userID, &group); err != nil {
		h.logger.Error("Failed to create group", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create group"})
		return
	}
	c.JSON(http.StatusOK, group)
}

// POST /api/goals
func (h *yearlyGoalHandler) CreateGoal(c *gin.Context) {
	userID := c.MustGet("userID").(string)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var goal models.YearlyGoal
	if err := c.ShouldBindJSON(&goal); err != nil {
		h.logger.Error("Invalid goal input", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Creating goal", slog.Any("goal", goal))

	if err := h.yearlyGoalService.CreateGoal(ctx, userID, &goal); err != nil {
		h.logger.Error("Failed to create goal", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, goal)
}

// PATCH /api/goals/:id/progress
type ProgressInput struct {
	CurrentStep int `json:"currentStep"`
}

func (h *yearlyGoalHandler) UpdateProgress(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	goalID := c.Param("id")

	var input ProgressInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Debug("Updating progress", slog.String("userID", userID), slog.String("goalID", goalID), slog.Int("currentStep", input.CurrentStep))

	if err := h.yearlyGoalService.UpdateProgress(c.Request.Context(), userID, goalID, input.CurrentStep); err != nil {
		h.logger.Error("Failed to update progress", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to update progress"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}
