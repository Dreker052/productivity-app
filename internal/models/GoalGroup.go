package models

type GoalGroup struct {
	ID     string       `json:"id"`
	Title  string       `json:"title"`
	Goals  []YearlyGoal `json:"goals"`
	UserID string       `json:"-"`
}
