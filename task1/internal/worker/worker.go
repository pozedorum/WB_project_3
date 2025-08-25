package worker

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"sync"
	"time"

	"github.com/pozedorum/WB_project_3/task1/internal/models"
	"github.com/pozedorum/WB_project_3/task1/internal/service"
	"github.com/pozedorum/wbf/zlog"
)

type Worker struct {
	service      *service.NotificationService
	wg           sync.WaitGroup
	shutdownChan chan struct{}
	semaphore    chan struct{}
}

func NewWorker(service *service.NotificationService) *Worker {
	return &Worker{
		service:      service,
		shutdownChan: make(chan struct{}),
		semaphore:    make(chan struct{}, 10),
	}
}

func (w *Worker) Start(ctx context.Context, queueName string) error {
	zlog.Logger.Info().Str("queue", queueName).Msg("Starting worker for queue")

	// Получаем канал сообщений из очереди
	messages, err := w.service.Consume(ctx, queueName)
	if err != nil {
		return err
	}

	w.wg.Add(1)
	go w.processMessages(ctx, messages)

	zlog.Logger.Info().Str("queue", queueName).Msg("Worker successfully started listening to queue")
	return nil
}

func (w *Worker) processMessages(ctx context.Context, messages <-chan []byte) {
	defer w.wg.Done()

	for {
		select {
		case <-w.shutdownChan:
			zlog.Logger.Info().Msg("Worker received shutdown signal")
			return

		case <-ctx.Done():
			zlog.Logger.Info().Msg("Worker context cancelled")
			return

		case msg, ok := <-messages:
			if !ok {
				zlog.Logger.Info().Msg("Messages channel closed")
				return
			}
			// не очень понимаю как правильно это реализовать,
			// в текущей реализации будут теряться сообщения при их большом количестве
			select {
			case w.semaphore <- struct{}{}:
				w.wg.Add(1)
				go func() {
					defer func() { <-w.semaphore }()
					defer w.wg.Done()
					w.processSingleMessage(ctx, msg)
				}()
			default:
				zlog.Logger.Warn().Msg("Worker busy, skipping message")
			}
		}
	}
}

func (w *Worker) processSingleMessage(ctx context.Context, messageData []byte) {
	// Логируем получение сообщения
	zlog.Logger.Debug().Str("message", string(messageData)).Msg("Received raw message from queue")

	// Декодируем base64 из []byte
	decodedData := make([]byte, base64.StdEncoding.DecodedLen(len(messageData)))
	n, err := base64.StdEncoding.Decode(decodedData, messageData)
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("Failed to decode base64 message")
		return
	}
	decodedData = decodedData[:n] // Обрезаем до фактической длины

	zlog.Logger.Debug().Str("message", string(decodedData)).Msg("Decoded message")

	// Парсим JSON
	var notification models.Notification
	if err := json.Unmarshal(decodedData, &notification); err != nil {
		zlog.Logger.Error().Err(err).Msg("Failed to unmarshal JSON message")
		return
	}

	zlog.Logger.Info().
		Str("notification_id", notification.ID).
		Str("user_id", notification.UserID).
		Msg("Processing notification")

	// Обрабатываем уведомление с retry логикой
	if err := w.processWithRetry(ctx, &notification); err != nil {
		zlog.Logger.Error().
			Err(err).
			Str("notification_id", notification.ID).
			Msg("Failed to process notification after retries")
	} else {
		zlog.Logger.Info().
			Str("notification_id", notification.ID).
			Msg("Successfully processed notification")
	}
}

func (w *Worker) processWithRetry(ctx context.Context, notification *models.Notification) error {
	retryStrategy := models.ConsumerStrategy
	var lastErr error

	for attempt := 1; attempt <= retryStrategy.Attempts; attempt++ {
		select {
		case <-w.shutdownChan:
			return lastErr
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Пытаемся обработать уведомление
			err := w.service.ProcessNotification(ctx, notification)
			if err == nil {
				return nil // Успех!
			}

			lastErr = err
			zlog.Logger.Warn().
				Err(err).
				Str("notification_id", notification.ID).
				Int("attempt", attempt).
				Int("max_attempts", retryStrategy.Attempts).
				Msg("Processing attempt failed")

			// Если это не последняя попытка, ждем перед повторной
			if attempt < retryStrategy.Attempts {
				delay := calculateBackoff(attempt, retryStrategy.Delay)
				zlog.Logger.Info().
					Str("notification_id", notification.ID).
					Dur("delay", delay).
					Int("next_attempt", attempt+1).
					Int("max_attempts", retryStrategy.Attempts).
					Msg("Retrying notification")

				select {
				case <-time.After(delay):
					continue
				case <-w.shutdownChan:
					return lastErr
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}
	}

	return lastErr
}

// calculateBackoff вычисляет время задержки с экспоненциальным откатом
func calculateBackoff(attempt int, baseDelay time.Duration) time.Duration {
	// Экспоненциальный backoff с максимальным ограничением
	backoff := baseDelay * time.Duration(1<<uint(attempt-1))

	// Максимальная задержка - 1 минута
	maxDelay := 1 * time.Minute
	if backoff > maxDelay {
		return maxDelay
	}
	return backoff
}

// Stop останавливает worker gracefully
func (w *Worker) Stop() {
	zlog.Logger.Info().Msg("Stopping worker gracefully...")
	close(w.shutdownChan)
	w.wg.Wait()
	zlog.Logger.Info().Msg("Worker stopped successfully")
}

// GetStatus возвращает статус worker (можно использовать для health checks)
func (w *Worker) GetStatus() map[string]interface{} {
	return map[string]interface{}{
		"active":     true,
		"goroutines": "1", // Базовая реализация
		"queue":      "notifications",
		"started_at": time.Now().Format(time.RFC3339),
		"processed":  "0", // Можно добавить счетчики при необходимости
	}
}
