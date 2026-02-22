package telegram

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/Dreker052/productivity-app/internal/service"
	"github.com/gin-gonic/gin"
)

type telegramHandler struct {
	telegramService service.TelegramService
	taskService     service.DailyTaskService
	logger          *slog.Logger
}

func NewTelegramHandler(telegramService service.TelegramService, taskService service.DailyTaskService, logger *slog.Logger) *telegramHandler {
	return &telegramHandler{
		telegramService: telegramService,
		taskService:     taskService,
		logger:          logger,
	}
}

// GET /api/telegram/link
func (h *telegramHandler) GetTelegramLink(c *gin.Context) {
	userID := c.MustGet("userID").(string)

	link, err := h.telegramService.GenerateLink(userID)
	if err != nil {
		h.logger.Error("Failed to generate link", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate link"})
		return
	}

	h.logger.Info("Generated Telegram link", slog.String("link", link))

	c.JSON(http.StatusOK, gin.H{"url": link})
}

// POST /api/share/telegram
func (h *telegramHandler) ShareDailyTasksToTelegram(c *gin.Context) {
	userID := c.MustGet("userID").(string)

	dateStr := c.Query("date")
	date := time.Now()
	if dateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, dateStr); err == nil {
			date = parsed
		}
	}

	tasks, err := h.taskService.GetByDate(c.Request.Context(), userID, date)
	if err != nil {
		h.logger.Error("Failed to fetch tasks", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks"})
		return
	}

	if len(tasks) == 0 {
		h.telegramService.SendDailyPlan(userID, "🤷‍♂️ *На этот день планов пока нет.*")
		c.JSON(http.StatusOK, gin.H{"status": "empty"})
		return
	}

	completedCount := 0
	for _, t := range tasks {
		if t.IsCompleted {
			completedCount++
		}
	}
	progressPercent := 0
	if len(tasks) > 0 {
		progressPercent = (completedCount * 100) / len(tasks)
	}

	var sb strings.Builder

	daysRu := map[string]string{
		"Monday": "Понедельник", "Tuesday": "Вторник", "Wednesday": "Среда",
		"Thursday": "Четверг", "Friday": "Пятница", "Saturday": "Суббота", "Sunday": "Воскресенье",
	}
	dayOfWeek := daysRu[date.Weekday().String()]
	fmt.Fprintf(&sb, "📅 <b>План на %s</b> (%s)\n", date.Format("02.01"), dayOfWeek)

	fmt.Fprintf(&sb, "🏆 <i>Прогресс: %d/%d (%d%%)</i>\n\n", completedCount, len(tasks), progressPercent)

	for _, t := range tasks {
		safeTitle := htmlEscape(t.Title)

		if t.IsCompleted {
			fmt.Fprintf(&sb, "<s>%s</s>\n", safeTitle)

		} else {
			fmt.Fprintf(&sb, "%s\n", safeTitle)
		}
	}

	err = h.telegramService.SendDailyPlan(userID, sb.String())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Telegram not linked"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "sent"})
}

func htmlEscape(text string) string {
	r := strings.NewReplacer("<", "&lt;", ">", "&gt;", "&", "&amp;")
	return r.Replace(text)
}
