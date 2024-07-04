package models

import "time"

type Task struct {
	UserID    int
	TaskID    int
	StartTime time.Time
	EndTime   time.Time
}
