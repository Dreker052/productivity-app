package user

import (
	"context"
	"log/slog"

	"github.com/Dreker052/productivity-app.git/internal/models"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepository struct {
	db     *pgxpool.Pool
	sb     sq.StatementBuilderType
	logger *slog.Logger
}

func NewUserRepository(db *pgxpool.Pool, logger *slog.Logger) *userRepository {
	return &userRepository{
		db:     db,
		logger: logger,
		sb:     sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	query, args, err := r.sb.Insert("users").
		Columns("id", "name", "email", "password").
		Values(user.ID, user.Name, user.Email, user.Password).ToSql()
	if err != nil {
		r.logger.Error("failed to create sql query", slog.String("error", err.Error()))
		return err
	}
	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		r.logger.Error("failed to execute sql query", slog.String("error", err.Error()))
		return err
	}

	r.logger.Info("user created successfully",
		slog.String("user_id", user.ID),
		slog.String("name", user.Name),
		slog.String("email", user.Email))

	return err
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query, args, err := r.sb.Select("id", "name", "email", "password").
		From("users").Where(sq.Eq{"email": email}).ToSql()
	if err != nil {
		r.logger.Error("failed to create sql query", slog.String("error", err.Error()))
		return nil, err
	}

	var user models.User
	err = r.db.QueryRow(ctx, query, args...).Scan(&user.ID, &user.Email, &user.Email, &user.Password)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	query, args, err := r.sb.Select("id", "name", "email", "password").
		From("users").Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		r.logger.Error("[UserRepo GetUserByID] failed to create sql query", slog.String("error", err.Error()))
		return nil, err
	}

	var user models.User
	err = r.db.QueryRow(ctx, query, args...).Scan(&user.ID, &user.Name, &user.Email, &user.Password)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}
