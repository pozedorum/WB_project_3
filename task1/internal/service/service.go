package service

import (
	"context"

	"github.com/pozedorum/WB_project_3/task1/internal/models"
	"github.com/pozedorum/WB_project_3/task1/internal/repository/postgres"
	"github.com/pozedorum/WB_project_3/task1/internal/repository/redis"
)

type NotificationService struct {
	repos *postgres.NotificationRepository
	cache *redis.NotificationCache
}

func New(pgRepo *postgres.NotificationRepository, redisRepo *redis.NotificationCache) *NotificationService {
	return &NotificationService{
		repos: pgRepo,
		cache: redisRepo,
	}
}

func (s *NotificationService) Create(ctx context.Context, req *models.CreateNotificationRequest) (*models.Notification, error) {
	return nil, nil
}

func (s *NotificationService) GetByID(ctx context.Context, id string) (*models.Notification, error) {
	return nil, nil
}

func (s *NotificationService) Delete(ctx context.Context, id string) error {
	return nil
}
