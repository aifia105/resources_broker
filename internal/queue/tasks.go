package queue

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

const TypeProvisionResource = "resource:provision"

type ProvisionPayload struct {
	TaskID       string                 `json:"task_id"`
	UserID       string                 `json:"user_id"`
	Action       string                 `json:"action"`
	Provider     string                 `json:"provider"`
	ResourceType string                 `json:"resource_type"`
	Payload      map[string]interface{} `json:"payload"`
}

func NewProvisionTask(payload ProvisionPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeProvisionResource, data), nil
}
