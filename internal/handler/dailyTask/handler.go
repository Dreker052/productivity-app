package dailytask

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/Dreker052/productivity-app.git/internal/models"
	"github.com/Dreker052/productivity-app.git/internal/service"
	"github.com/gin-gonic/gin"
)

type dailyTaskHandler struct {
	taskService service.DailyTaskService
	logger      *slog.Logger
}

func NewDailyTaskHandler(s service.DailyTaskService, l *slog.Logger) *dailyTaskHandler {
	return &dailyTaskHandler{
		taskService: s,
		logger:      l,
	}
}

// POST /api/tasks
func (h *dailyTaskHandler) Create(c *gin.Context) {
	userID := c.MustGet("userID").(string)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var task models.DailyTask
	if err := c.ShouldBindJSON(&task); err != nil {
		h.logger.Error("[TaskHandler Create] Failed to bind JSON", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task.UserID = userID

	if err := h.taskService.Create(ctx, &task); err != nil {
		h.logger.Error("[TaskHandler Create] Failed to create task", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	h.logger.Info("[TaskHandler Create] Task created successfully", slog.String("taskID", task.ID))

	c.JSON(http.StatusOK, task)
}

// GET /api/tasks?date=2024-01-19T...
func (h *dailyTaskHandler) GetTasks(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	dateStr := c.Query("date")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	date := time.Now()
	if dateStr != "" {
		parsed, err := time.Parse(time.RFC3339, dateStr)
		if err == nil {
			date = parsed
		}
	}

	tasks, err := h.taskService.GetByDate(ctx, userID, date)
	if err != nil {
		h.logger.Error("[TaskHandler GetTasks] Failed to get tasks", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tasks"})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

// PATCH /api/tasks/:id/toggle
func (h *dailyTaskHandler) ToggleTask(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	taskID := c.Param("id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	err := h.taskService.ToggleStatus(ctx, userID, taskID)
	if err != nil {
		h.logger.Error("Toggle failed", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to update task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}
