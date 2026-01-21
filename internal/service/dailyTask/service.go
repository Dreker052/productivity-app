package dailytask

import (
	"context"
	"log/slog"
	"time"

	"github.com/Dreker052/productivity-app.git/internal/models"
	"github.com/Dreker052/productivity-app.git/internal/repository"
)

type dailyTaskService struct {
	repo   repository.DailyTaskRepository
	logger *slog.Logger
}

func NewDailyTaskService(repo repository.DailyTaskRepository, logger *slog.Logger) *dailyTaskService {
	return &dailyTaskService{
		repo:   repo,
		logger: logger,
	}
}

func (s *dailyTaskService) Create(ctx context.Context, task *models.DailyTask) error {
	return s.repo.Create(ctx, task)
}

func (s *dailyTaskService) GetByDate(ctx context.Context, userID string, date time.Time) ([]*models.DailyTask, error) {
	return s.repo.GetByDate(ctx, userID, date)
}

func (s *dailyTaskService) ToggleStatus(ctx context.Context, userID string, taskID string) error {
	return s.repo.ToggleStatus(ctx, userID, taskID)
}
