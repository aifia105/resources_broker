package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"resources_broker/internal/model"
	"resources_broker/internal/queue"
	"time"

	"github.com/hibiken/asynq"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const taskCollectionName = "Task"

var (
	ErrInvalidTaskID = errors.New("invalid task id")
	ErrTaskNotFound  = errors.New("task not found")
	ErrInvalidTask   = errors.New("invalid task")
)

type TaskService struct {
	DB          *mongo.Database
	QueueClient *asynq.Client
}

func NewTaskService(db *mongo.Database, queueClient *asynq.Client) *TaskService {
	return &TaskService{DB: db, QueueClient: queueClient}
}

func (s *TaskService) CreateTask(ctx context.Context, request *model.Task) (*model.Task, error) {
	now := time.Now().UTC()
	task := &model.Task{
		UserID:       request.UserID,
		Status:       model.StatusPending,
		Action:       request.Action,
		ResourceType: request.ResourceType,
		Provider:     request.Provider,
		Payload:      request.Payload,
		RetryCount:   0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if task.Payload == nil {
		task.Payload = map[string]interface{}{}
	}

	result, err := s.DB.Collection(taskCollectionName).InsertOne(ctx, task)
	if err != nil {
		return nil, err
	}
	createdID, err := insertedTaskID(result.InsertedID)
	if err != nil {
		return nil, err
	}
	task.ID = createdID

	job, err := queue.NewProvisionTask(queue.ProvisionPayload{
		TaskID:       task.ID.Hex(),
		UserID:       task.UserID,
		Action:       task.Action,
		Provider:     task.Provider,
		ResourceType: task.ResourceType,
		Payload:      task.Payload,
	})
	if err != nil {
		s.DB.Collection(taskCollectionName).DeleteOne(ctx, bson.M{"_id": task.ID})
		return nil, fmt.Errorf("failed to create provisioning task: %w", err)
	}

	if _, err := s.QueueClient.Enqueue(job,
		asynq.MaxRetry(4),
		asynq.Timeout(5*time.Minute),
		asynq.Queue("default")); err != nil {
		s.DB.Collection(taskCollectionName).DeleteOne(ctx, bson.M{"_id": task.ID})
		return nil, fmt.Errorf("failed to enqueue provisioning task: %w", err)
	}

	task.Status = model.StatusQueued
	task.UpdatedAt = time.Now().UTC()
	if _, err := s.DB.Collection(taskCollectionName).UpdateOne(
		ctx,
		bson.M{"_id": task.ID},
		bson.M{"$set": bson.M{"status": task.Status, "updated_at": task.UpdatedAt}},
	); err != nil {
		log.Printf("warn: task %s enqueued but status update failed: %v", task.ID.Hex(), err)
	}
	return task, nil
}

func (s *TaskService) GetTaskByID(ctx context.Context, userID string, taskID string) (*model.Task, error) {
	filter, err := taskFilter(userID, taskID)
	if err != nil {
		return nil, err
	}

	var task model.Task
	err = s.DB.Collection(taskCollectionName).FindOne(ctx, filter).Decode(&task)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}
	return &task, nil
}

func (s *TaskService) GetTasks(ctx context.Context, userID string) ([]model.Task, error) {
	filter := bson.M{"user_id": userID}
	cursor, err := s.DB.Collection(taskCollectionName).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var tasks []model.Task
	if err := cursor.All(ctx, &tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (s *TaskService) UpdateTaskStatus(ctx context.Context, taskID string, status model.Status, errorMessage *string) error {
	taskObjectID, err := bson.ObjectIDFromHex(taskID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidTaskID, err)
	}

	result, err := s.DB.Collection(taskCollectionName).UpdateOne(ctx,
		bson.M{"_id": taskObjectID},
		bson.M{"$set": bson.M{"status": status,
			"error_message": errorMessage,
			"updated_at":    time.Now().UTC()}})

	if err != nil {
		return err
	}
	if result.ModifiedCount == 0 {
		return ErrTaskNotFound
	}
	return nil
}

func taskFilter(userID string, taskID string) (bson.M, error) {
	taskObjectID, err := bson.ObjectIDFromHex(taskID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidTaskID, err)
	}

	return bson.M{
		"_id":     taskObjectID,
		"user_id": userID,
	}, nil
}

func insertedTaskID(insertedID any) (bson.ObjectID, error) {
	switch value := insertedID.(type) {
	case bson.ObjectID:
		return value, nil
	case string:
		return bson.ObjectIDFromHex(value)
	default:
		return bson.ObjectID{}, fmt.Errorf("unexpected inserted id type %T", insertedID)
	}
}
