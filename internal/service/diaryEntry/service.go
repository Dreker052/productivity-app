package dailytask

import (
	"context"
	"log/slog"
	"time"

	"github.com/Dreker052/productivity-app.git/internal/models"
	"github.com/Dreker052/productivity-app.git/internal/repository"
)

type diaryEntryService struct {
	diaryEntryRepo repository.DiaryEntryRepository
	logger         *slog.Logger
}

func NewDiaryEntryService(diaryEntryRepo repository.DiaryEntryRepository, logger *slog.Logger) *diaryEntryService {
	return &diaryEntryService{
		diaryEntryRepo: diaryEntryRepo,
		logger:         logger,
	}
}

func (s *diaryEntryService) Save(ctx context.Context, entry *models.DiaryEntry) error {
	return s.diaryEntryRepo.Save(ctx, entry)
}

func (s *diaryEntryService) GetByDate(ctx context.Context, userID string, date time.Time) (*models.DiaryEntry, error) {
	return s.diaryEntryRepo.GetByDate(ctx, userID, date)
}
