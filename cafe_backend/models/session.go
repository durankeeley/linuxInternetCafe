package models

import "time"

type Session struct {
	ID         int       `json:"id"`
	ComputerID int       `json:"computer_id"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	Status     string    `json:"status"` // "active", "expired"
}
