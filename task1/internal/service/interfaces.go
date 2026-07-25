package service

import (
	"context"
	"time"

	"github.com/pozedorum/WB_project_3/task1/internal/models"
)

type NotificationService interface {
	Create(ctx context.Context, req *models.CreateNotificationRequest) (*models.Notification, error)
	GetByID(ctx context.Context, id string) (*models.Notification, error)
	Delete(ctx context.Context, id string) error
	Consume(ctx context.Context, queueName string) (<-chan []byte, error)
	ProcessNotification(ctx context.Context, notification *models.Notification) error
	ProcessNotificationData(ctx context.Context, data []byte) error
}

// Repository интерфейс для работы с данными
type Repository interface {
	CreateNotification(ctx context.Context, n *models.Notification) error
	GetByID(ctx context.Context, id string) (*models.Notification, error)
	UpdateNotificationStatus(ctx context.Context, id, status string) error
	DeleteNotification(ctx context.Context, id string) error
}

// Cache интерфейс для кэширования
type Cache interface {
	Set(ctx context.Context, key string, value interface{}) error
	Get(ctx context.Context, key string) (*models.Notification, error)
	Ping(ctx context.Context) (string, error)
	Close() error
}

// Queue интерфейс для работы с очередями
type Queue interface {
	PublishWithDelay(ctx context.Context, queueName string, message interface{}, delay time.Duration) error
	Consume(ctx context.Context, queueName string) (<-chan []byte, error)
	Close() error
}

// Notifier интерфейс для отправки уведомлений
type Notifier interface {
	Send(notification *models.Notification) error
	GetChannel() string
}
