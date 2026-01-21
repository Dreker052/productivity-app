package dailytask

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Dreker052/productivity-app.git/internal/models"
	"github.com/Masterminds/squirrel"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type dailyTaskRepository struct {
	db     *pgxpool.Pool
	sb     sq.StatementBuilderType
	logger *slog.Logger
}

func NewDailyTaskRepository(db *pgxpool.Pool, logger *slog.Logger) *dailyTaskRepository {
	return &dailyTaskRepository{
		db:     db,
		sb:     sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
		logger: logger,
	}
}

func (r *dailyTaskRepository) Create(ctx context.Context, task *models.DailyTask) error {
	query, args, err := r.sb.Insert("daily_tasks").
		Columns("id", "title", "is_completed", "date", "user_id").
		Values(task.ID, task.Title, task.IsCompleted, task.Date, task.UserID).
		ToSql()

	if err != nil {
		r.logger.Error("[TaskRepository CreateTask] Failed to build create task query", slog.String("error", err.Error()))
		return err
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		r.logger.Error("[TaskRepository CreateTask] Failed to create task", slog.String("error", err.Error()))
		return err
	}

	r.logger.Info("[TaskRepository CreateTask] Task created successfully", slog.String("taskID", task.ID))
	return nil
}

func (r *dailyTaskRepository) GetByDate(ctx context.Context, userID string, date time.Time) ([]*models.DailyTask, error) {
	query, args, err := r.sb.Select("id", "title", "is_completed", "date").
		From("daily_tasks").
		Where(sq.And{
			sq.Eq{"user_id": userID},
			sq.Eq{"date": date},
		}).
		ToSql()

	if err != nil {
		r.logger.Error("[TaskRepository GetByDate] Failed to build query", slog.String("error", err.Error()))
		return nil, err
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		r.logger.Error("[TaskRepository GetByDate] Failed to fetch tasks", slog.String("error", err.Error()))
		return nil, err
	}
	defer rows.Close()

	tasks := []*models.DailyTask{}

	for rows.Next() {
		var task models.DailyTask
		if err := rows.Scan(&task.ID, &task.Title, &task.IsCompleted, &task.Date); err != nil {
			r.logger.Error("[TaskRepository GetByDate] Failed to scan task", slog.String("error", err.Error()))
			return nil, err
		}
		tasks = append(tasks, &task)
	}

	return tasks, nil
}

func (r *dailyTaskRepository) ToggleStatus(ctx context.Context, userID string, taskID string) error {

	sql, args, err := r.sb.Update("daily_tasks").
		Set("is_completed", squirrel.Expr("NOT is_completed")).
		Where(squirrel.Eq{
			"id":      taskID,
			"user_id": userID,
		}).
		ToSql()

	if err != nil {
		r.logger.Error("[TaskRepository ToggleStatus] Failed to build update query", slog.String("error", err.Error()))
		return err
	}

	res, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		r.logger.Error("[TaskRepository ToggleStatus] Failed to toggle task", slog.String("id", taskID), slog.String("error", err.Error()))
		return err
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("task not found or access denied")
	}

	return nil
}
