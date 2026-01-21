package models

import "time"

type DailyTask struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	IsCompleted bool      `json:"isCompleted"`
	Date        time.Time `json:"date"`
	UserID      string    `json:"userId"`
}
