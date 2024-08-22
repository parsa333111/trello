package database

import (
	"time"
)

type Workspace struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Task struct {
	ID            uint      `json:"id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Status        string    `json:"status"`
	EstimatedTime int       `json:"estimated_time"`
	ActualTime    int       `json:"actual_time"`
	DueDate       time.Time `json:"due_date"`
	Priority      int       `json:"priority"`
	WorkspaceID   uint      `json:"workspace_id"`
	AssigneeID    uint      `json:"assignee_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	ImageURL      string    `json:"image_url"`
}

type Subtask struct {
	ID          uint      `json:"id"`
	TaskID      uint      `json:"task_id"`
	Title       string    `json:"title"`
	IsCompleted string    `json:"is_completed"`
	AssigneeID  uint      `json:"assignee_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type User struct {
	ID           uint      `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash []byte    `json:"password_hash"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UserWorkspaceRole struct {
	ID          uint      `json:"id"`
	UserID      uint      `json:"user_id"`
	WorkspaceID uint      `json:"workspace_id"`
	Role        string    `json:"role"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Comment struct {
	ID     uint   `json:"id"`
	TaskID uint   `json:"task_id"`
	UserID uint   `json:"user_id"`
	Text   string `json:"text"`
}

type Watch struct {
	TaskID uint `json:"task_id"`
	UserID uint `json:"user_id"`
}

type WatchStatus struct {
	Status string `json:"status"`
}

type ImageStatus struct {
	Status string `json:"status"`
}
