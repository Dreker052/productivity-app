package diaryEntry

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/Dreker052/productivity-app/internal/models"
	"github.com/Dreker052/productivity-app/internal/service"
	"github.com/gin-gonic/gin"
)

type diaryEntryHandler struct {
	diaryEntryService service.DiaryEntryService
	logger            *slog.Logger
}

func NewDiaryEntryHandler(diaryEntryService service.DiaryEntryService, logger *slog.Logger) *diaryEntryHandler {
	return &diaryEntryHandler{diaryEntryService: diaryEntryService, logger: logger}
}

// POST /api/diary
func (h *diaryEntryHandler) Save(c *gin.Context) {
	userID := c.MustGet("userID").(string)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var entry models.DiaryEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		h.logger.Error("[DiaryEntryHandler SaveDiaryEntry] Failed to bind diary entry JSON", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entry.UserID = userID

	if err := h.diaryEntryService.Save(ctx, &entry); err != nil {
		h.logger.Error("Failed to save diary", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save entry"})
		return
	}

	h.logger.Info("Diary entry saved successfully", slog.String("entryID", entry.ID))

	c.JSON(http.StatusOK, entry)
}

// GET /api/diary?date=2024-01-18T...
func (h *diaryEntryHandler) GetByDate(c *gin.Context) {
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

	entry, err := h.diaryEntryService.GetByDate(ctx, userID, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if entry == nil {
		c.JSON(http.StatusOK, nil)
		return
	}

	h.logger.Info("Diary entry retrieved successfully", slog.String("entryID", entry.ID))

	c.JSON(http.StatusOK, entry)
}
