package handlers

import (
	"errors"
	"net/http"
	"resources_broker/internal/jwt"
	"resources_broker/internal/model"
	"resources_broker/internal/service"

	"github.com/gin-gonic/gin"
)

type TaskHandler struct {
	service *service.TaskService
}

func NewTaskHandler(service *service.TaskService) *TaskHandler {
	return &TaskHandler{service: service}
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
	userID, err := jwt.ExtractTokenID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var request model.CreateTaskRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := h.service.CreateTask(c.Request.Context(), &model.Task{
		UserID:       userID.Hex(),
		Action:       request.Action,
		ResourceType: request.ResourceType,
		Provider:     request.Provider,
		Payload:      request.Payload,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create task"})
		return
	}

	c.JSON(http.StatusCreated, toTaskResponse(task))
}

func (h *TaskHandler) GetTask(c *gin.Context) {
	userID, err := jwt.ExtractTokenID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	task, err := h.service.GetTaskByID(c.Request.Context(), userID.Hex(), c.Param("id"))
	if err != nil {
		status := http.StatusInternalServerError
		message := "failed to fetch task"
		if errors.Is(err, service.ErrInvalidTaskID) {
			status = http.StatusBadRequest
			message = "invalid task id"
		} else if errors.Is(err, service.ErrTaskNotFound) {
			status = http.StatusNotFound
			message = "task not found"
		}
		c.JSON(status, gin.H{"error": message})
		return
	}

	c.JSON(http.StatusOK, toTaskResponse(task))
}

func (h *TaskHandler) ListTasks(c *gin.Context) {
	userID, err := jwt.ExtractTokenID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	tasks, err := h.service.GetTasks(c.Request.Context(), userID.Hex())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list tasks"})
		return
	}

	c.JSON(http.StatusOK, toTaskResponses(tasks))
}

func toTaskResponse(task *model.Task) model.TaskResponse {
	return model.TaskResponse{
		ID:           task.ID.Hex(),
		UserID:       task.UserID,
		Status:       task.Status,
		CreatedAt:    task.CreatedAt,
		UpdatedAt:    task.UpdatedAt,
		ErrorMessage: task.ErrorMessage,
	}
}

func toTaskResponses(tasks []model.Task) []model.TaskResponse {
	responses := make([]model.TaskResponse, len(tasks))
	for i, task := range tasks {
		responses[i] = toTaskResponse(&task)
	}
	return responses
}
