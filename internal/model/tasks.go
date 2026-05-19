package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusQueued    Status = "queued"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusRetrying  Status = "retrying"
)

type Task struct {
	ID           bson.ObjectID          `json:"id" bson:"_id,omitempty"`
	UserID       string                 `json:"user_id" bson:"user_id"`
	Status       Status                 `json:"status" bson:"status"`
	Action       string                 `json:"action" bson:"action"`
	ResourceType string                 `json:"resource_type" bson:"resource_type"`
	Provider     string                 `json:"provider" bson:"provider"`
	Payload      map[string]interface{} `json:"payload" bson:"payload"`
	ResourceID   *string                `json:"resource_id,omitempty" bson:"resource_id,omitempty"`
	RetryCount   int                    `json:"retry_count" bson:"retry_count"`
	CreatedAt    time.Time              `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" bson:"updated_at"`
	ErrorMessage *string                `json:"error_message,omitempty" bson:"error_message,omitempty"`
}

type TaskResponse struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Status       Status    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	ErrorMessage *string   `json:"error_message,omitempty"`
}

type CreateTaskRequest struct {
	Action       string                 `json:"action" binding:"required"`
	ResourceType string                 `json:"resource_type" binding:"required"`
	Provider     string                 `json:"provider" binding:"required"`
	Payload      map[string]interface{} `json:"payload"`
}
