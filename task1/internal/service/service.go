package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pozedorum/WB_project_3/task1/internal/models"
	"github.com/pozedorum/wbf/zlog"
)

type NotificationService struct {
	repo      Repository
	cache     Cache
	queue     Queue
	notifiers map[string]Notifier // map[channel]Notifier
}

func NewNotificationService(repo Repository, cache Cache, queue Queue, notifiers []Notifier) *NotificationService {
	notifierMap := make(map[string]Notifier)
	for _, notifier := range notifiers {
		notifierMap[notifier.GetChannel()] = notifier
	}

	return &NotificationService{
		repo:      repo,
		cache:     cache,
		queue:     queue,
		notifiers: notifierMap,
	}
}

func (s *NotificationService) Create(ctx context.Context, req *models.CreateNotificationRequest) (*models.Notification, error) {
	// Валидация
	if err := s.validateRequest(req); err != nil {
		return nil, err
	}

	// Проверяем поддержку канала
	if _, exists := s.notifiers[req.Channel]; !exists {
		return nil, fmt.Errorf("unsupported channel: %s", req.Channel)
	}

	// Создаем уведомление
	notification := &models.Notification{
		ID:        uuid.New().String(),
		UserID:    req.UserID,
		Message:   req.Message,
		Channel:   req.Channel,
		SendAt:    req.SendAt,
		Status:    models.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Сохраняем в репозиторий
	if err := s.repo.CreateNotification(ctx, notification); err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	// Кэшируем
	if err := s.cache.Set(ctx, notification.ID, notification); err != nil {
		zlog.Logger.Warn().Err(err).Str("notification_id", notification.ID).Msg("Failed to cache notification")
	}

	// Отправляем в очередь
	if err := s.publishToQueue(ctx, notification); err != nil {
		// Откатываем изменения при ошибке
		if err := s.repo.UpdateNotificationStatus(ctx, notification.ID, models.StatusFailed); err != nil {
			zlog.Logger.Error().Err(err).Str("notification_id", notification.ID).Msg("Failed to update status after queue error")
		}
		return nil, fmt.Errorf("failed to publish to queue: %w", err)
	}

	return notification, nil
}

func (s *NotificationService) GetByID(ctx context.Context, id string) (*models.Notification, error) {
	// Пробуем из кэша
	if cached, err := s.cache.Get(ctx, id); err == nil && cached != nil {
		if cached == nil {
			// Запись была удалена
			return nil, nil
		}
		return cached, nil
	}

	// Ищем в репозитории
	notification, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	if notification == nil {
		return nil, nil
	}

	// Обновляем кэш асинхронно
	notificationCopy := *notification
	go func() {
		if err := s.cache.Set(context.Background(), id, &notificationCopy); err != nil {
			zlog.Logger.Error().Err(err).Str("notification_id", id).Msg("Failed to update cache")
		}
	}()

	return notification, nil
}

func (s *NotificationService) Delete(ctx context.Context, id string) error {
	notification, err := s.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get notification: %w", err)
	}

	if notification == nil {
		return fmt.Errorf("notification not found")
	}

	if notification.Status == models.StatusSent {
		return fmt.Errorf("cannot delete sent notification")
	}

	if notification.Status == models.StatusCanceled {
		return fmt.Errorf("notification already canceled")
	}
	// Удаляем из репозитория
	if err := s.repo.DeleteNotification(ctx, id); err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}
	// Устанавливаем новый статус в redis
	if err := s.cache.Set(ctx, id, nil); err != nil {
		zlog.Logger.Warn().Err(err).Str("notification_id", id).Msg("Failed to update cache after deletion")
	}
	return nil
}

// Функция для воркера
func (s *NotificationService) Consume(ctx context.Context, queueName string) (<-chan []byte, error) {
	return s.queue.Consume(ctx, queueName)
}

func (s *NotificationService) ProcessNotificationData(ctx context.Context, data []byte) error {
	var notification models.Notification
	if err := json.Unmarshal(data, &notification); err != nil {
		return fmt.Errorf("failed to unmarshal notification: %w", err)
	}
	return s.ProcessNotification(ctx, &notification)
}

func (s *NotificationService) ProcessNotification(ctx context.Context, notification *models.Notification) error {
	zlog.Logger.Info().Str("notification_id", notification.ID).Msg("Processing notification")

	// Проверяем актуальность уведомления
	current, err := s.GetByID(ctx, notification.ID)
	if err != nil {
		return fmt.Errorf("failed to get current notification state: %w", err)
	}

	if current == nil || current.Status != models.StatusPending {
		return fmt.Errorf("notification is no longer pending")
	}

	// Отправляем уведомление
	notifier, exists := s.notifiers[notification.Channel]
	if !exists {
		return fmt.Errorf("no notifier for channel: %s", notification.Channel)
	}

	if err := notifier.Send(current); err != nil {
		// Обновляем статус на failed
		if updateErr := s.repo.UpdateNotificationStatus(ctx, notification.ID, models.StatusFailed); updateErr != nil {
			zlog.Logger.Error().Err(updateErr).Str("notification_id", notification.ID).Msg("Failed to update status after send error")
		}
		return fmt.Errorf("failed to send notification: %w", err)
	}

	// Обновляем статус на sent
	if err := s.repo.UpdateNotificationStatus(ctx, notification.ID, models.StatusSent); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	// Обновляем кэш
	current.Status = models.StatusSent
	if err := s.cache.Set(ctx, notification.ID, current); err != nil {
		zlog.Logger.Error().Err(err).Str("notification_id", notification.ID).Msg("Failed to update cache")
	}

	zlog.Logger.Info().Str("notification_id", notification.ID).Msg("Notification processed successfully")
	return nil
}

// Вспомогательные методы
func (s *NotificationService) validateRequest(req *models.CreateNotificationRequest) error {
	if req.UserID == "" {
		return fmt.Errorf("user_id is required")
	}
	if req.Message == "" {
		return fmt.Errorf("message is required")
	}
	if req.Channel == "" {
		return fmt.Errorf("channel is required")
	}
	if req.SendAt.Before(time.Now().Add(1 * time.Minute)) {
		return fmt.Errorf("send_at must be at least 1 minute in the future")
	}
	return nil
}

func (s *NotificationService) publishToQueue(ctx context.Context, notification *models.Notification) error {
	delay := time.Until(notification.SendAt)
	if delay < 0 {
		delay = 0
	}

	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}
	return s.queue.PublishWithDelay(ctx, "notifications", data, delay)
}
