package models

import "time"

type DiaryEntry struct {
	ID     string    `json:"id"`
	Text   string    `json:"text"`
	Date   time.Time `json:"date"`
	UserID string    `json:"-"`
}
