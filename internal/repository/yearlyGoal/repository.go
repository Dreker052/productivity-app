package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Dreker052/productivity-app/internal/models"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type yearlyGoalRepository struct {
	db     *pgxpool.Pool
	sb     sq.StatementBuilderType
	logger *slog.Logger
}

func NewYearlyGoalRepository(db *pgxpool.Pool, logger *slog.Logger) *yearlyGoalRepository {
	return &yearlyGoalRepository{
		db:     db,
		sb:     sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
		logger: logger,
	}
}

func (r *yearlyGoalRepository) CreateGoalGroup(ctx context.Context, group *models.GoalGroup) error {
	query, args, err := r.sb.Insert("goal_groups").
		Columns("id", "title", "user_id").
		Values(group.ID, group.Title, group.UserID).
		ToSql()

	if err != nil {
		r.logger.Error("Failed to build SQL for creating goal group", slog.String("error", err.Error()))
		return err
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to create goal group", slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (r *yearlyGoalRepository) CreateYearlyGoal(ctx context.Context, userID string, goal *models.YearlyGoal) error {

	checkQuery, checkArgs, _ := r.sb.Select("1").
		From("goal_groups").
		Where(sq.And{
			sq.Eq{"id": goal.GoalGroupID},
			sq.Eq{"user_id": userID},
		}).
		Limit(1).
		ToSql()

	var exists int
	err := r.db.QueryRow(ctx, checkQuery, checkArgs...).Scan(&exists)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("access denied: group not found or belongs to another user")
		}
		return err
	}

	query, args, err := r.sb.Insert("yearly_goals").
		Columns("id", "title", "total_steps", "current_step", "goal_group_id").
		Values(goal.ID, goal.Title, goal.TotalSteps, goal.CurrentStep, goal.GoalGroupID).
		ToSql()

	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to create goal", slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (r *yearlyGoalRepository) GetAllGoalGroups(ctx context.Context, userID string) ([]*models.GoalGroup, error) {

	queryGroups, argsGroups, _ := r.sb.Select("id", "title").
		From("goal_groups").
		Where(sq.Eq{"user_id": userID}).
		ToSql()

	rows, err := r.db.Query(ctx, queryGroups, argsGroups...)
	if err != nil {
		r.logger.Error("Failed to fetch groups", slog.String("error", err.Error()))
		return nil, err
	}
	defer rows.Close()

	var groups []*models.GoalGroup
	groupsMap := make(map[string]*models.GoalGroup)
	var groupIDs []string

	for rows.Next() {
		var g models.GoalGroup
		if err := rows.Scan(&g.ID, &g.Title); err != nil {
			return nil, err
		}
		g.Goals = []models.YearlyGoal{}

		newGroup := g
		groups = append(groups, &newGroup)
		groupIDs = append(groupIDs, g.ID)
	}
	rows.Close()

	if len(groups) == 0 {
		return []*models.GoalGroup{}, nil
	}

	for _, grp := range groups {
		groupsMap[grp.ID] = grp
	}

	queryGoals, argsGoals, _ := r.sb.Select("id", "title", "total_steps", "current_step", "goal_group_id").
		From("yearly_goals").
		Where(sq.Eq{"goal_group_id": groupIDs}).
		ToSql()

	rowsGoals, err := r.db.Query(ctx, queryGoals, argsGoals...)
	if err != nil {
		r.logger.Error("Failed to fetch goals", slog.String("error", err.Error()))
		return nil, err
	}
	defer rowsGoals.Close()

	for rowsGoals.Next() {
		var goal models.YearlyGoal
		if err := rowsGoals.Scan(&goal.ID, &goal.Title, &goal.TotalSteps, &goal.CurrentStep, &goal.GoalGroupID); err != nil {
			return nil, err
		}

		if group, exists := groupsMap[goal.GoalGroupID]; exists {
			group.Goals = append(group.Goals, goal)
		}
	}

	return groups, nil
}

func (r *yearlyGoalRepository) UpdateProgress(ctx context.Context, userID string, goalID string, currentStep int) error {

	qb := sq.StatementBuilder.PlaceholderFormat(sq.Question)

	subQuerySql, subArgs, err := qb.Select("id").
		From("goal_groups").
		Where(sq.Eq{"user_id": userID}).
		ToSql()

	if err != nil {
		r.logger.Error("Failed to build subquery", slog.String("error", err.Error()))
		return err
	}

	query := qb.Update("yearly_goals").
		Set("current_step", sq.Expr("?::integer", currentStep)).
		Where(sq.And{
			sq.Eq{"id": goalID},

			sq.Expr("goal_group_id IN ("+subQuerySql+")", subArgs...),
		})

	sql, args, err := query.PlaceholderFormat(sq.Dollar).ToSql()

	if err != nil {
		r.logger.Error("Failed to build final SQL", slog.String("error", err.Error()))
		return err
	}

	cmd, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		r.logger.Error("SQL Exec Error", slog.String("error", err.Error()))
		return err
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("goal not found or access denied")
	}

	return nil

}
