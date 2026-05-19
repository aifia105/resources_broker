package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"resources_broker/internal/model"
	"resources_broker/internal/queue"
	"resources_broker/internal/service"

	"github.com/hibiken/asynq"
)

type ProvisionHandler struct {
	provisionService *service.ProvisionResourceService
	taskService      *service.TaskService
}

func NewProvisionHandler(provisionService *service.ProvisionResourceService, taskService *service.TaskService) *ProvisionHandler {
	return &ProvisionHandler{provisionService: provisionService, taskService: taskService}
}

func (h *ProvisionHandler) HandleProvisionTask(ctx context.Context, job *asynq.Task) error {
	var payload queue.ProvisionPayload
	if err := json.Unmarshal(job.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task payload: %w", err)
	}

	if err := h.taskService.UpdateTaskStatus(ctx, payload.TaskID, model.StatusRunning, nil); err != nil {
		return fmt.Errorf("failed to update task status to running: %w", err)
	}

	if err := h.provisionService.ProvisionResource(ctx, payload); err != nil {
		message := err.Error()
		if updateErr := h.taskService.UpdateTaskStatus(ctx, payload.TaskID, model.StatusFailed, &message); updateErr != nil {
			log.Printf("error: failed to update task status to failed: %v (original error: %v)", updateErr, err)
		}
		return nil
	}

	if err := h.taskService.UpdateTaskStatus(ctx, payload.TaskID, model.StatusCompleted, nil); err != nil {
		log.Printf("error: failed to update task status to completed: %v", err)
	}

	return nil
}
