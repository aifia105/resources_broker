package service

import (
	"context"

	"resources_broker/internal/queue"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type ProvisionResourceService struct {
	DB *mongo.Database
}

func NewProvisionResourceService(db *mongo.Database) *ProvisionResourceService {
	return &ProvisionResourceService{DB: db}
}

func (s *ProvisionResourceService) ProvisionResource(ctx context.Context, payload queue.ProvisionPayload) error {
	_ = ctx
	_ = payload

	return nil
}
