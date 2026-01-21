package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type telegramRepository struct {
	db     *pgxpool.Pool
	sb     sq.StatementBuilderType
	logger *slog.Logger
}

func NewTelegramRepository(db *pgxpool.Pool, logger *slog.Logger) *telegramRepository {
	return &telegramRepository{
		db:     db,
		sb:     sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
		logger: logger,
	}
}

func (r *telegramRepository) SaveLinkToken(ctx context.Context, token string, userID string, ttl time.Duration) error {
	expiresAt := time.Now().Add(ttl)

	query, args, err := r.sb.Insert("telegram_link_tokens").
		Columns("token", "user_id", "expires_at").
		Values(token, userID, expiresAt).
		ToSql()

	if err != nil {
		r.logger.Error("Failed to build SQL for saving link token", slog.String("error", err.Error()))
		return err
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to save telegram link token", slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (r *telegramRepository) GetUserIDByToken(ctx context.Context, token string) (string, error) {
	query, args, err := r.sb.Select("user_id").
		From("telegram_link_tokens").
		Where(sq.Eq{"token": token}).
		Where(sq.Gt{"expires_at": time.Now()}).
		ToSql()

	if err != nil {
		r.logger.Error("Failed to build SQL for getting user by token", slog.String("error", err.Error()))
		return "", err
	}

	var userID string
	err = r.db.QueryRow(ctx, query, args...).Scan(&userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("token not found or expired")
		}
		r.logger.Error("Failed to get user by token", slog.String("error", err.Error()))
		return "", err
	}

	delQuery, delArgs, _ := r.sb.Delete("telegram_link_tokens").Where(sq.Eq{"token": token}).ToSql()
	_, _ = r.db.Exec(ctx, delQuery, delArgs...)

	return userID, nil
}

func (r *telegramRepository) SaveIntegration(ctx context.Context, userID string, chatID int64, username string) error {
	query, args, err := r.sb.Insert("telegram_integrations").
		Columns("user_id", "chat_id", "username").
		Values(userID, chatID, username).
		Suffix("ON CONFLICT (user_id) DO UPDATE SET chat_id = EXCLUDED.chat_id, username = EXCLUDED.username, created_at = NOW()").
		ToSql()

	if err != nil {
		r.logger.Error("Failed to build SQL for saving telegram integration", slog.String("error", err.Error()))
		return err
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to save telegram integration", slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (r *telegramRepository) GetChatID(ctx context.Context, userID string) (int64, error) {
	query, args, err := r.sb.Select("chat_id").
		From("telegram_integrations").
		Where(sq.Eq{"user_id": userID}).
		ToSql()

	if err != nil {
		r.logger.Error("Failed to build SQL for getting chat_id", slog.String("error", err.Error()))
		return 0, err
	}

	var chatID int64
	err = r.db.QueryRow(ctx, query, args...).Scan(&chatID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, fmt.Errorf("telegram integration not found")
		}
		r.logger.Error("Failed to get chat_id", slog.String("error", err.Error()))
		return 0, err
	}

	return chatID, nil
}
