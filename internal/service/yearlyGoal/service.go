package yearlygoal

import (
	"context"
	"log/slog"

	"github.com/Dreker052/productivity-app.git/internal/models"
	"github.com/Dreker052/productivity-app.git/internal/repository"
)

type yearlyGoalService struct {
	yearlyGoalRepo repository.YearlyGoalRepository
	logger         *slog.Logger
}

func NewYearlyGoalService(yearlyGoalRepo repository.YearlyGoalRepository, logger *slog.Logger) *yearlyGoalService {
	return &yearlyGoalService{
		yearlyGoalRepo: yearlyGoalRepo,
		logger:         logger,
	}
}

func (s *yearlyGoalService) CreateGroup(ctx context.Context, userID string, group *models.GoalGroup) error {
	group.UserID = userID
	return s.yearlyGoalRepo.CreateGoalGroup(ctx, group)
}

func (s *yearlyGoalService) CreateGoal(ctx context.Context, userID string, goal *models.YearlyGoal) error {
	return s.yearlyGoalRepo.CreateYearlyGoal(ctx, userID, goal)
}

func (s *yearlyGoalService) GetGroups(ctx context.Context, userID string) ([]*models.GoalGroup, error) {
	return s.yearlyGoalRepo.GetAllGoalGroups(ctx, userID)
}

func (s *yearlyGoalService) UpdateProgress(ctx context.Context, userID, goalID string, step int) error {
	return s.yearlyGoalRepo.UpdateProgress(ctx, userID, goalID, step)
}
