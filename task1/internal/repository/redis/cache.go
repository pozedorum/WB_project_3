package redis

import (
	"context"
	"encoding/json"

	"github.com/pozedorum/WB_project_3/task1/internal/models"
	"github.com/pozedorum/WB_project_3/task1/internal/service"
	"github.com/pozedorum/wbf/redis"
	"github.com/pozedorum/wbf/zlog"
)

type NotificationCache struct {
	client *redis.Client
}

func NewNotificationRepository(addr, password string, db int) *NotificationCache {
	zlog.Logger.Info().Str("address", addr).Int("db", db).Msg("Creating Redis notification cache")
	return &NotificationCache{client: redis.New(addr, password, db)}
}

func (ns *NotificationCache) Set(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		zlog.Logger.Error().Err(err).Str("key", key).Msg("Failed to marshal value for cache")
		return err
	}

	err = ns.client.SetWithRetry(ctx, models.StandartStrategy, "notification-"+key, data)
	if err != nil {
		zlog.Logger.Error().Err(err).Str("key", key).Msg("Failed to set value in cache")
	} else {
		zlog.Logger.Debug().Str("key", key).Msg("Value set in cache")
	}
	return err
}

func (ns *NotificationCache) Get(ctx context.Context, key string) (*models.Notification, error) {
	data, err := ns.client.GetWithRetry(ctx, models.StandartStrategy, "notification-"+key)
	if err != nil {
		zlog.Logger.Error().Err(err).Str("key", key).Msg("Failed to get value from cache")
		return nil, err
	}

	var n models.Notification
	err = json.Unmarshal([]byte(data), &n)
	if err != nil {
		zlog.Logger.Error().Err(err).Str("key", key).Msg("Failed to unmarshal value from cache")
		return nil, err
	}

	zlog.Logger.Debug().Str("key", key).Msg("Value retrieved from cache")
	return &n, nil
}

func (ns *NotificationCache) Close() error {
	zlog.Logger.Info().Msg("Closing Redis connection")
	return ns.client.Close()
}

func (ns *NotificationCache) Ping(ctx context.Context) (string, error) {
	result, err := ns.client.Ping(ctx).Result()
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("Redis ping failed")
	} else {
		zlog.Logger.Debug().Str("result", result).Msg("Redis ping successful")
	}
	return result, err
}

var _ service.Cache = (*NotificationCache)(nil)
