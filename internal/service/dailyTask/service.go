package dailytask

import (
	"context"
	"log/slog"
	"time"

	"github.com/Dreker052/productivity-app.git/internal/models"
	"github.com/Dreker052/productivity-app.git/internal/repository"
)

type dailyTaskService struct {
	dailyTaskRepo repository.DailyTaskRepository
	logger        *slog.Logger
}

func NewDailyTaskService(dailyTaskRepo repository.DailyTaskRepository, logger *slog.Logger) *dailyTaskService {
	return &dailyTaskService{
		dailyTaskRepo: dailyTaskRepo,
		logger:        logger,
	}
}

func (s *dailyTaskService) Create(ctx context.Context, task *models.DailyTask) error {
	return s.dailyTaskRepo.Create(ctx, task)
}

func (s *dailyTaskService) GetByDate(ctx context.Context, userID string, date time.Time) ([]*models.DailyTask, error) {
	return s.dailyTaskRepo.GetByDate(ctx, userID, date)
}

func (s *dailyTaskService) ToggleStatus(ctx context.Context, userID string, taskID string) error {
	return s.dailyTaskRepo.ToggleStatus(ctx, userID, taskID)
}

func (s *dailyTaskService) Delete(ctx context.Context, userID string, taskID string) error {
	return s.dailyTaskRepo.Delete(ctx, userID, taskID)
}

func (s *dailyTaskService) Update(ctx context.Context, userID, taskID, newTitle string) error {
	return s.dailyTaskRepo.Update(ctx, userID, taskID, newTitle)
}
