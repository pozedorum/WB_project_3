package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pozedorum/WB_project_3/task1/internal/models"
	"github.com/pozedorum/WB_project_3/task1/pkg/rabbitmq"
)

type NotificationWorker struct {
	channel   *rabbitmq.Channel
	consumer  *rabbitmq.Consumer
	service   *service.NotificationService
}

func NewNotificationWorker(rabbitURL string, service *service.NotificationService) (*NotificationWorker, error) {
	conn, err := rabbitmq.Connect(rabbitURL, 5, 2*time.Second)
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	// Конфигурация consumer'а
	config := rabbitmq.NewConsumerConfig("notifications")
	config.AutoAck = false // Ручное подтверждение!
	config.Exclusive = false

	consumer := rabbitmq.NewConsumer(channel, config)

	return &NotificationWorker{
		channel:  channel,
		consumer: consumer,
		service:  service,
	}, nil
}

func (w *NotificationWorker) Start(ctx context.Context) error {
	// Создаем канал для сообщений
	msgChan := make(chan []byte, 100)
	
	// Запускаем потребление
	go func() {
		if err := w.consumer.Consume(msgChan); err != nil {
			log.Printf("Consume error: %v", err)
		}
	}()

	log.Println("Worker started. Waiting for messages...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Worker stopping...")
			return nil
			
		case msg, ok := <-msgChan:
			if !ok {
				log.Println("Message channel closed")
				return nil
			}
			
			// Обрабатываем сообщение
			if err := w.processMessage(msg); err != nil {
				log.Printf("Failed to process message: %v", err)
			}
		}
	}
}

func (w *NotificationWorker) processMessage(data []byte) error {
	// Парсим уведомление
	var notification models.Notification
	if err := json.Unmarshal(data, &notification); err != nil {
		return fmt.Errorf("failed to unmarshal notification: %w", err)
	}

	log.Printf("Processing notification: %s (scheduled for: %s)", 
		notification.ID, notification.SendAt.Format(time.RFC3339))

	// Проверяем актуальность (не отменено ли)
	current, err := w.service.GetByID(context.Background(), notification.ID)
	if err != nil {
		return fmt.Errorf("failed to check notification: %w", err)
	}

	if current == nil || current.Status != models.StatusPending {
		log.Printf("Notification %s is no longer pending. Skipping.", notification.ID)
		return nil
	}

	// Отправляем уведомление
	if err := w.service.SendNotification(context.Background(), &notification); err != nil {
		// Помечаем как failed в БД
		if updateErr := w.service.MarkAsFailed(context.Background(), notification.ID); updateErr != nil {
			log.Printf("Failed to mark as failed: %v", updateErr)
		}
		return fmt.Errorf("failed to send notification: %w", err)
	}

	// Помечаем как sent в БД
	if err := w.service.MarkAsSent(context.Background(), notification.ID); err != nil {
		return fmt.Errorf("failed to mark as sent: %w", err)
	}

	log.Printf("Notification %s processed successfully", notification.ID)
	return nil
}

func main() {
	// Инициализация сервиса и пр.
	service := // ... ваш код инициализации
	
	worker, err := NewNotificationWorker("amqp://localhost:5672", service)
	if err != nil {
		log.Fatal("Failed to create worker:", err)
	}

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), 
		os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := worker.Start(ctx); err != nil {
		log.Fatal("Worker failed:", err)
	}
}