package repository

import (
	"context"
	"time"

	"github.com/Dreker052/productivity-app.git/internal/models"
)

type DailyTaskRepository interface {
	Create(ctx context.Context, task *models.DailyTask) error
	GetByDate(ctx context.Context, userId string, date time.Time) ([]*models.DailyTask, error)
	ToggleStatus(ctx context.Context, userId string, taskId string) error
}

type DiaryEntryRepository interface {
	Save(ctx context.Context, entry *models.DiaryEntry) error
	GetByDate(ctx context.Context, userId string, date time.Time) (*models.DiaryEntry, error)
}

type YearlyGoalRepository interface {
	CreateGoalGroup(ctx context.Context, group *models.GoalGroup) error
	CreateYearlyGoal(ctx context.Context, userID string, goal *models.YearlyGoal) error
	GetAllGoalGroups(ctx context.Context, userID string) ([]*models.GoalGroup, error)
	UpdateProgress(ctx context.Context, userID string, goalID string, currentStep int) error
}

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	ChangeName(ctx context.Context, userID string, newName string) error
	ChangeEmail(ctx context.Context, userID, email string) error
	ChangePassword(ctx context.Context, userID, newHash string) error
	SaveVerificationToken(ctx context.Context, token, userID string, ttl time.Duration) error
	VerifyEmail(ctx context.Context, verificationToken string) error
}

type TelegramRepository interface {
	SaveLinkToken(ctx context.Context, token string, userID string, ttl time.Duration) error
	GetUserIDByToken(ctx context.Context, token string) (string, error)
	SaveIntegration(ctx context.Context, userID string, chatID int64, username string) error
	GetChatID(ctx context.Context, userID string) (int64, error)
}
