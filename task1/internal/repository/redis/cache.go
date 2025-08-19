package redis

import (
	"context"
	"encoding/json"

	"github.com/pozedorum/WB_project_3/task1/internal/models"
	"github.com/pozedorum/wbf/redis"
)

type NotificationCache struct {
	client *redis.Client
}

func NewNotificationRepository(addr, password string, db int) *NotificationCache {
	return &NotificationCache{client: redis.New(addr, password, db)}
}

func (ns *NotificationCache) Set(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return ns.client.SetWithRetry(ctx, models.StandartStrategy, "notification-"+key, data)
}

func (ns *NotificationCache) Get(ctx context.Context, key string) (*models.Notification, error) {
	data, err := ns.client.GetWithRetry(ctx, models.StandartStrategy, "notification-"+key)
	if err != nil {
		return nil, err
	}
	var n models.Notification
	err = json.Unmarshal([]byte(data), &n)
	if err != nil {
		return nil, err
	}
	return &n, nil
}
