package models

type YearlyGoal struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	TotalSteps  int    `json:"totalSteps"`
	CurrentStep int    `json:"currentStep"`
	GoalGroupID string `json:"goalGroupId"`
}
