package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/pozedorum/WB_project_3/task4/internal/models"
	"github.com/pozedorum/WB_project_3/task4/internal/service"
	"github.com/pozedorum/wbf/zlog"

	// в пакете фреймворка нет обёртки над kafka.Message, так что нужно использовать стандартную библиотеку

	wbfKafka "github.com/wb-go/wbf/kafka"
)

// KafkaTaskConsumer потребитель задач из Kafka
type KafkaTaskConsumer struct {
	consumer  *wbfKafka.Consumer
	processor service.ImageProcessor
	storage   service.Storage
	repo      service.Repository
	topic     string
	groupID   string
}

// NewKafkaTaskConsumer создает нового потребителя задач
func NewKafkaTaskConsumer(
	consumer *wbfKafka.Consumer,
	processor service.ImageProcessor,
	storage service.Storage,
	repo service.Repository,
	topic, groupID string,
) *KafkaTaskConsumer {
	return &KafkaTaskConsumer{
		consumer:  consumer,
		processor: processor,
		storage:   storage,
		repo:      repo,
		topic:     topic,
		groupID:   groupID,
	}
}

// StartConsuming запускает процесс потребления сообщений из Kafka.
func (c *KafkaTaskConsumer) StartConsuming(ctx context.Context, workersCount int) {

	zlog.Logger.Info().Str("topic", c.topic).Str("group_id", c.groupID).Msg("Kafka consumer started")

	var wg sync.WaitGroup
	wg.Add(workersCount)
	taskCh := make(chan Message, workersCount*2) // Буферизованный канал для задач
	// Запускаем пул воркеров
	for i := 0; i < workersCount; i++ {
		go func(workerID int) {
			defer wg.Done() // Уменьшаем счетчик при завершении воркера
			c.worker(ctx, taskCh, workerID)
		}(i)
	}
	go func() {
		<-ctx.Done() // Ждем сигнала завершения
		zlog.Logger.Info().Msg("Shutdown signal received, closing task channel...")
		close(taskCh) // Закрываем канал, чтобы воркеры могли завершиться
	}()

	// Основной цикл обработки
	for {
		select {
		case <-ctx.Done():
			// Контекст отменен, выходим из цикла
			zlog.Logger.Info().Msg("Stopping message consumption...")
			break
		default:
			msg, err := c.consumer.FetchWithRetry(ctx, models.ProduserConsumerStrategy)
			if err != nil {
				if ctx.Err() != nil {
					// Ошибка из-за отмененного контекста - нормальная ситуация при shutdown
					zlog.Logger.Debug().Msg("Context cancelled, stopping fetch")
					break
				}
				zlog.Logger.Error().Err(err).Msg("Failed to fetch message from Kafka")
				time.Sleep(1 * time.Second)
				continue
			}

			// Пытаемся отправить сообщение в канал с таймаутом
			select {
			case taskCh <- msg:
				// Сообщение успешно отправлено в канал
			case <-ctx.Done():
				// Не удалось отправить из-за shutdown, можно вернуть сообщение?
				zlog.Logger.Warn().Msg("Shutdown during message routing, message might be lost")
				break
			}
		}

		// Проверяем, нужно ли выходить из цикла
		if ctx.Err() != nil {
			break
		}
	}

	// ЖДЕМ завершения всех воркеров
	zlog.Logger.Info().Msg("Waiting for workers to finish...")
	wg.Wait()
	zlog.Logger.Info().Msg("All workers finished, consumer shutdown complete")

}

// processMessage обрабатывает одно сообщение из Kafka.
func (c *KafkaTaskConsumer) processMessage(ctx context.Context, msg Message) error {
	// 1. Декодируем сообщение
	var processingMsg models.ProcessingMessage
	err := json.Unmarshal(msg.Value, &processingMsg)
	if err != nil {
		zlog.Logger.Error().Err(err).Bytes("message_value", msg.Value).Msg("Failed to unmarshal Kafka message")
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}
	taskID := processingMsg.TaskID
	zlog.Logger.Info().Str("task_id", taskID).Msg("Starting to process image task")

	// 2. Обновляем статус в БД на "processing"
	err = c.repo.UpdateImageStatus(ctx, taskID, "processing")
	if err != nil {
		zlog.Logger.Error().Err(err).Str("task_id", taskID).Msg("Failed to update status to 'processing'")
		return err
	}

	// 3. Получаем оригинальное изображение из хранилища
	originalData, err := c.storage.Get(ctx, "originals/"+taskID)
	if err != nil {
		zlog.Logger.Error().Err(err).Str("task_id", taskID).Msg("Failed to get original image from storage")
		_ = c.repo.UpdateImageStatus(ctx, taskID, "failed")
		return fmt.Errorf("failed to get original image: %w", err)
	}

	// 4. Обрабатываем изображение
	result, err := c.processor.ProcessImage(originalData, processingMsg.Options)
	if err != nil {
		zlog.Logger.Error().Err(err).Str("task_id", taskID).Msg("Image processing failed")
		// Пытаемся обновить статус на "failed", но не прерываем выполнение из-за этой ошибки
		_ = c.repo.UpdateImageStatus(ctx, taskID, "failed")
		return fmt.Errorf("image processing failed: %w", err)
	}

	// 5. Сохраняем обработанное изображение в хранилище
	processedFilename := "processed/" + taskID
	_, err = c.storage.Save(ctx, result.ProcessedData, processedFilename)
	if err != nil {
		zlog.Logger.Error().Err(err).Str("task_id", taskID).Msg("Failed to save processed image")
		_ = c.repo.UpdateImageStatus(ctx, taskID, "failed")
		return fmt.Errorf("failed to save processed image: %w", err)
	}

	// 6. Обновляем метаданные в БД на "completed"
	metadata, err := c.repo.GetImageMetadata(ctx, taskID)
	if err != nil {
		zlog.Logger.Error().Err(err).Str("task_id", taskID).Msg("Failed to get metadata for update")
		return fmt.Errorf("failed to get metadata: %w", err)
	}

	metadata.Status = "completed"
	metadata.ProcessedAt = time.Now()
	metadata.Width = result.Width
	metadata.Height = result.Height
	metadata.Size = result.Size
	metadata.Format = result.Format
	metadata.ProcessedName = processedFilename

	err = c.repo.SaveImageMetadata(ctx, metadata)
	if err != nil {
		zlog.Logger.Error().Err(err).Str("task_id", taskID).Msg("Failed to update metadata to 'completed'")
		return fmt.Errorf("failed to update metadata: %w", err)
	}

	zlog.Logger.Info().
		Str("task_id", taskID).
		Int("width", result.Width).
		Int("height", result.Height).
		Str("format", result.Format).
		Msg("Image task processed successfully")

	return nil // Успешная обработка, ошибок нет
}

func (c *KafkaTaskConsumer) worker(ctx context.Context, taskCh <-chan Message, workerId int) {
	for msg := range taskCh {
		// 2. Обрабатываем сообщение
		processingErr := c.processMessage(ctx, msg)

		// 3. Если обработка прошла успешно, подтверждаем сообщение (commit)
		if processingErr == nil {
			var commitErr error
			for i := 0; i < 3; i++ {
				commitErr = c.consumer.Commit(ctx, msg)
				if commitErr == nil {
					zlog.Logger.Debug().Str("offset", fmt.Sprint(msg.Offset)).Msg("Message committed successfully")
					break
				}
				zlog.Logger.Warn().Err(commitErr).Str("offset", fmt.Sprint(msg.Offset)).Msg("Commit failed, retrying...")
				time.Sleep(time.Duration(i+1) * 500 * time.Millisecond)
			}
			if commitErr != nil {
				zlog.Logger.Error().Err(commitErr).Str("offset", fmt.Sprint(msg.Offset)).Msg("Failed to commit message after retries")
			}
		} else {
			// Если обработка не удалась, НЕ коммитим сообщение.
			// Оно будет обработано снова другим воркером или после перезапуска.
			zlog.Logger.Error().Err(processingErr).Str("offset", fmt.Sprint(msg.Offset)).Msg("Message processing failed, skipping commit")
		}
	}
	zlog.Logger.Debug().Msg("Worker shutting down")
}
