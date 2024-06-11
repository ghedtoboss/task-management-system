package models

type UserStats struct {
	UserID         int `json:"user_id"`
	TotalTasks     int `json:"total_tasks"`
	CompletedTasks int `json:"completed_tasks"`
	PendingTasks   int `json:"pending_tasks"`
}
