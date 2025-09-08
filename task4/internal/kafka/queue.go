package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/pozedorum/WB_project_3/task4/internal/models"
	"github.com/pozedorum/wbf/zlog"

	//"github.com/segmentio/kafka-go"
	wbfKafka "github.com/wb-go/wbf/kafka"
)

func InitKafka() {
	// Устанавливаем уровень логирования для Kafka библиотеки
	os.Setenv("KAFKA_GO_LOG_LEVEL", "warn")
}

// KafkaImageQueue реализация ImageQueue через wbf/kafka
type KafkaImageQueue struct {
	producer *wbfKafka.Producer
	topic    string
}

// NewKafkaImageQueue создает новую Kafka очередь для изображений
func NewKafkaImageQueue(producer *wbfKafka.Producer, topic string) *KafkaImageQueue {
	return &KafkaImageQueue{
		producer: producer,
		topic:    topic,
	}
}

// PublishImageTask публикует задачу на обработку изображения
func (q *KafkaImageQueue) PublishImageTask(ctx context.Context, task *models.ProcessingTask) error {
	message := models.ProcessingMessage{
		TaskID:      task.TaskID,
		Options:     task.Options,
		CallbackURL: task.CallbackURL,
		Timestamp:   time.Now(),
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = q.producer.SendWithRetry(ctx, models.ProduserConsumerStrategy, []byte(task.TaskID), messageBytes)
	if err != nil {
		return fmt.Errorf("failed to publish task to Kafka: %w", err)
	}

	zlog.Logger.Info().
		Str("task_id", task.TaskID).
		Str("topic", q.topic).
		Msg("Image task published to Kafka")

	return nil
}

// HealthCheck проверяет соединение с Kafka
func (q *KafkaImageQueue) HealthCheck(ctx context.Context) error {
	// Простая проверка - отправляем тестовое сообщение
	testMessage := models.ProcessingMessage{
		TaskID:    "healthcheck",
		Timestamp: time.Now(),
	}

	testBytes, _ := json.Marshal(testMessage)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return q.producer.Send(ctx, []byte("healthcheck"), testBytes)
}

// Close закрывает соединение
func (q *KafkaImageQueue) Close() error {
	return q.producer.Close()
}
