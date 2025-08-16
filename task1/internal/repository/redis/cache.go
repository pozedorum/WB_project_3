package redis

import (
	"time"

	"github.com/pozedorum/wbf/redis"
)

type NotificationCache struct {
	client    *redis.Client
	timeLimit time.Duration
}

func NewNotificationRepository(addr, password string, db int, timeLimit time.Duration) *NotificationCache {
	return &NotificationCache{client: redis.New(addr, password, db), timeLimit: timeLimit}
}
