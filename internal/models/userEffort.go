package models

import (
	"sort"
)

type UserEffort struct {
	UserID  int
	TaskID  int
	Hours   int
	Minutes int
}

func CalculateUserEffort(tasks []Task) []UserEffort {
	var userEfforts []UserEffort

	for _, task := range tasks {
		var userEffort UserEffort
		userEffort.UserID = task.UserID
		userEffort.TaskID = task.TaskID

		duration := task.EndTime.Sub(task.StartTime)

		userEffort.Hours = int(duration.Hours())
		userEffort.Minutes = int(duration.Minutes())
		userEfforts = append(userEfforts, userEffort)
	}
	return userEfforts
}

func SortUserEfforts(userEfforts []UserEffort) {
	if len(userEfforts) > 1 {
		sort.Slice(userEfforts, func(i, j int) bool {
			return userEfforts[i].Hours*60+userEfforts[i].Minutes > userEfforts[j].Hours*60+userEfforts[j].Minutes
		})
	}
}
