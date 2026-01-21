package diaryentry

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/Dreker052/productivity-app.git/internal/models"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type diaryEntryRepository struct {
	db     *pgxpool.Pool
	sb     squirrel.StatementBuilderType
	logger *slog.Logger
}

func NewDiaryEntryRepository(db *pgxpool.Pool, logger *slog.Logger) *diaryEntryRepository {
	return &diaryEntryRepository{
		db:     db,
		sb:     squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger: logger,
	}
}

func (r *diaryEntryRepository) Save(ctx context.Context, entry *models.DiaryEntry) error {
	query, args, err := r.sb.Insert("diary_entries").
		Columns("id", "text", "date", "user_id").
		Values(entry.ID, entry.Text, entry.Date, entry.UserID).
		Suffix("ON CONFLICT (id) DO UPDATE SET text = EXCLUDED.text").
		ToSql()

	if err != nil {
		r.logger.Error("[DiaryEntryRepo Save] Failed to create SQL query", slog.String("error", err.Error()))
		return err
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to save diary entry", slog.String("error", err.Error()))
		return err
	}

	r.logger.Info("Diary entry saved successfully", slog.String("entry_id", entry.ID))

	return nil
}

func (r *diaryEntryRepository) GetByDate(ctx context.Context, userID string, date time.Time) (*models.DiaryEntry, error) {
	query, args, err := r.sb.Select("id", "text", "date", "user_id").
		From("diary_entries").
		Where(squirrel.Eq{
			"user_id": userID,
			"date":    date,
		}).
		Limit(1).
		ToSql()

	if err != nil {
		r.logger.Error("[DiaryEntryRepo GetByDate] Failed to create SQL query", slog.String("error", err.Error()))
		return nil, err
	}

	var entry models.DiaryEntry
	err = r.db.QueryRow(ctx, query, args...).Scan(&entry.ID, &entry.Text, &entry.Date, &entry.UserID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		r.logger.Error("[DiaryEntryRepo GetByDate] Failed to fetch diary entry", slog.String("error", err.Error()))
		return nil, err
	}

	return &entry, nil
}
