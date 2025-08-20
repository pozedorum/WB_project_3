package service

import (
	"context"
	"time"

	"github.com/pozedorum/WB_project_3/task1/internal/models"
)

// Repository интерфейс для работы с данными
type Repository interface {
	CreateNotification(ctx context.Context, n *models.Notification) error
	GetByID(ctx context.Context, id string) (*models.Notification, error)
	UpdateNotificationStatus(ctx context.Context, id, status string) error
	DeleteNotification(ctx context.Context, id string) error
	GetPendingNotifications(ctx context.Context) ([]*models.Notification, error)
}

// Cache интерфейс для кэширования
type Cache interface {
	Set(ctx context.Context, key string, value interface{}) error
	Get(ctx context.Context, key string) (*models.Notification, error)
	Delete(ctx context.Context, key string) error
}

// Queue интерфейс для работы с очередями
type Queue interface {
	PublishWithDelay(ctx context.Context, queueName string, message interface{}, delay time.Duration) error
	Consume(ctx context.Context, queueName string) (<-chan []byte, error)
}

// Notifier интерфейс для отправки уведомлений
type Notifier interface {
	Send(notification *models.Notification) error
	GetChannel() string
}
