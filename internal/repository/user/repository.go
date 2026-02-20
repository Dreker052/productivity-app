package user

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/Dreker052/productivity-app/internal/models"
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
	query, args, err := r.sb.Select("id", "name", "email", "password", "is_verified").
		From("users").Where(sq.Eq{"email": email}).ToSql()
	if err != nil {
		r.logger.Error("failed to create sql query", slog.String("error", err.Error()))
		return nil, err
	}

	var user models.User
	err = r.db.QueryRow(ctx, query, args...).Scan(&user.ID, &user.Email, &user.Email, &user.Password, &user.IsVerified)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	query, args, err := r.sb.Select("id", "name", "email", "password", "is_verified").
		From("users").Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		r.logger.Error("[UserRepo GetUserByID] failed to create sql query", slog.String("error", err.Error()))
		return nil, err
	}

	var user models.User
	err = r.db.QueryRow(ctx, query, args...).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.IsVerified)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) ChangeName(ctx context.Context, userID, newName string) error {
	sql, args, err := r.sb.Update("users").
		Set("name", newName).
		Where(sq.Eq{"id": userID}).
		ToSql()
	if err != nil {
		r.logger.Error("[UserRepo ChangeName] failed to create update name query", slog.String("error", err.Error()))
		return err
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		r.logger.Error("[UserRepo ChangeName] failed to update name", slog.String("error", err.Error()))
		return err
	}

	r.logger.Info("[UserRepo ChangeName] name successfully updated", slog.String("userID", userID))

	return nil
}

func (r *userRepository) ChangeEmail(ctx context.Context, userID, newEmail string) error {
	sql, args, err := r.sb.Update("users").
		Set("email", newEmail).
		Where(sq.Eq{"id": userID}).
		ToSql()
	if err != nil {
		r.logger.Error("[UserRepo ChangeName] failed to create change email query", slog.String("error", err.Error()))
		return err
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		r.logger.Error("[UserRepo ChangeName] failed to change email", slog.String("error", err.Error()))
		return err
	}

	r.logger.Info("[UserRepo ChangeName] email successfully changed", slog.String("userID", userID))

	return nil
}

func (r *userRepository) ChangePassword(ctx context.Context, userID, newHash string) error {
	sql, args, err := r.sb.Update("users").
		Set("password", newHash).
		Where(sq.Eq{"id": userID}).
		ToSql()
	if err != nil {
		r.logger.Error("[UserRepo ChangePassword] failed to create change password query", slog.String("error", err.Error()))
		return err
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		r.logger.Error("[UserRepo ChangePassword] failed to change password", slog.String("error", err.Error()))
		return err
	}

	r.logger.Info("[UserRepo ChangePassword] password successfully changed", slog.String("userID", userID))

	return nil
}

func (r *userRepository) SaveVerificationToken(ctx context.Context, token, userID string, ttl time.Duration) error {
	sql, args, err := r.sb.Insert("verification_tokens").
		Columns("token", "user_id", "expires_at").
		Values(token, userID, time.Now().Add(ttl)).
		ToSql()
	if err != nil {
		r.logger.Error("[UserRepo SaveVerificationToken] failed to create insert query", slog.String("error", err.Error()))
		return err
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		r.logger.Error("[UserRepo SaveVerificationToken] failed to save verification token", slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (r *userRepository) VerifyEmail(ctx context.Context, verificationToken string) error {

	tx, err := r.db.Begin(ctx)
	if err != nil {
		r.logger.Error("[UserRepo VerifyEmail] failed to begin transaction", slog.String("error", err.Error()))
		return err
	}

	defer tx.Rollback(ctx)

	tokenSql, tokenArgs, err := r.sb.Select("user_id").
		From("verification_tokens").
		Where(sq.Eq{"token": verificationToken}).
		Where(sq.Expr("expires_at > NOW()")).
		ToSql()

	if err != nil {
		r.logger.Error("[UserRepo VerifyEmail] failed to create find token query", slog.String("error", err.Error()))
		return err
	}

	var userID string
	if err = tx.QueryRow(ctx, tokenSql, tokenArgs...).Scan(&userID); err != nil {
		if err == pgx.ErrNoRows {
			return errors.New("invalid or expired verification token")
		}
		r.logger.Error("[UserRepo VerifyEmail] failed to check verification token", slog.String("error", err.Error()))
		return err
	}

	updateSql, updateArgs, err := r.sb.Update("users").
		Set("is_verified", true).
		Where(sq.Eq{"id": userID}).
		ToSql()
	if err != nil {
		r.logger.Error("[UserRepo VerifyEmail] failed to create update user query", slog.String("error", err.Error()))
		return err
	}

	_, err = tx.Exec(ctx, updateSql, updateArgs...)
	if err != nil {
		r.logger.Error("[UserRepo VerifyEmail] failed to verify user email", slog.String("error", err.Error()))
		return err
	}

	deleteSql, deleteArgs, err := r.sb.Delete("verification_tokens").
		Where(sq.Eq{"token": verificationToken}).
		ToSql()
	if err != nil {
		r.logger.Error("[UserRepo VerifyEmail] failed to create delete token query", slog.String("error", err.Error()))
		return err
	}

	_, err = tx.Exec(ctx, deleteSql, deleteArgs...)
	if err != nil {
		r.logger.Error("[UserRepo VerifyEmail] failed to delete verification token", slog.String("error", err.Error()))
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		r.logger.Error("[UserRepo VerifyEmail] failed to commit transaction", slog.String("error", err.Error()))
		return err
	}

	r.logger.Info("[UserRepo VerifyEmail] user email successfully verified", slog.String("userID", userID))

	return nil
}
