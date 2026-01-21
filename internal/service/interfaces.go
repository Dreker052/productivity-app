package service

import (
	"context"
	"errors"
	"time"

	"github.com/Dreker052/productivity-app.git/internal/models"
)

var ErrUserAlreadyExists = errors.New("user already exists")

type DailyTaskService interface {
	Create(ctx context.Context, task *models.DailyTask) error
	GetByDate(ctx context.Context, userId string, date time.Time) ([]*models.DailyTask, error)
	ToggleStatus(ctx context.Context, userId string, taskId string) error
}

type DiaryEntryService interface {
	Save(ctx context.Context, entry *models.DiaryEntry) error
	GetByDate(ctx context.Context, userId string, date time.Time) (*models.DiaryEntry, error)
}

type YearlyGoalService interface {
	CreateGroup(ctx context.Context, userID string, group *models.GoalGroup) error
	CreateGoal(ctx context.Context, userID string, goal *models.YearlyGoal) error
	GetGroups(ctx context.Context, userID string) ([]*models.GoalGroup, error)
	UpdateProgress(ctx context.Context, userID, goalID string, currentStep int) error
}

type UserService interface {
	Create(ctx context.Context, user *models.User) error
	GetByEmail(ctx context.Context, email string) (*models.User, error)
}

type AuthService interface {
	Login(ctx context.Context, email, password string) (string, string, error)
	Register(ctx context.Context, user *models.User) (string, string, error)
	RefreshTokens(ctx context.Context, refreshToken string) (string, string, error)
}

type TelegramService interface {
	Start(ctx context.Context) error
	GenerateLink(userID string) (string, error)
	SendDailyPlan(userID string, planText string) error
	Stop()
}
