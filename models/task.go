package models

import "time"

type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	StartDate   time.Time `json:"start_date"`
	DueDate     time.Time `json:"due_date"`
	UserID      int       `json:"user_id"`
	AssignedTo  int       `json:"assigned_to"`
}
