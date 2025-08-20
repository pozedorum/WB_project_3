package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/pozedorum/wbf/rabbitmq"
)

type RabbitMQAdapter struct {
	conn      *rabbitmq.Connection
	channel   *rabbitmq.Channel
	exchange  *rabbitmq.Exchange
	publisher *rabbitmq.Publisher
	queue     *rabbitmq.Queue
}

func NewRabbitMQAdapter(url string) (*RabbitMQAdapter, error) {
	// Устанавливаем соединение
	conn, err := rabbitmq.Connect(url, 5, 2*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Создаем обычный direct exchange (не нужен delayed plugin)
	exchange := rabbitmq.NewExchange("notifications", "direct")
	exchange.Durable = true

	// Объявляем exchange в RabbitMQ
	if err := exchange.BindToChannel(channel); err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Создаем основную очередь для уведомлений
	qm := rabbitmq.NewQueueManager(channel)
	queue, err := qm.DeclareQueue("notifications", rabbitmq.QueueConfig{
		Durable:    true,
		AutoDelete: false,
	})
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	// Привязываем очередь к exchange
	err = channel.QueueBind(
		queue.Name,
		"notifications", // routing key
		exchange.Name,
		false,
		nil,
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	// Создаем publisher
	publisher := rabbitmq.NewPublisher(channel, "notifications")

	return &RabbitMQAdapter{
		conn:      conn,
		channel:   channel,
		exchange:  exchange,
		publisher: publisher,
		queue:     queue,
	}, nil
}

func (a *RabbitMQAdapter) PublishWithDelay(ctx context.Context, queueName string, message interface{}, delay time.Duration) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Вычисляем задержку в миллисекундах
	delayMs := delay.Milliseconds()
	if delayMs < 0 {
		delayMs = 0
	}

	// Публикуем с TTL (Expiration)
	err = a.publisher.PublishWithOptions(
		data,
		"application/json",
		rabbitmq.PublishOptions{
			Expiration: time.Duration(delayMs) * time.Millisecond,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish with delay: %w", err)
	}

	log.Printf("Published message with %v delay (TTL: %dms)", delay, delayMs)
	return nil
}

func (a *RabbitMQAdapter) Consume(ctx context.Context, queueName string) (<-chan []byte, error) {
	// Создаем канал для сообщений
	messages := make(chan []byte, 100)

	// Запускаем потребитель
	consumer := rabbitmq.NewConsumer(a.channel, a.queue.Name)

	deliveries, err := consumer.Consume(rabbitmq.ConsumeOptions{
		AutoAck:   false, // Ручное подтверждение
		Exclusive: false,
		NoLocal:   false,
		NoWait:    false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start consumer: %w", err)
	}

	// Обрабатываем сообщения в горутине
	go func() {
		defer close(messages)

		for {
			select {
			case <-ctx.Done():
				return
			case delivery, ok := <-deliveries:
				if !ok {
					return
				}

				// Передаем сообщение в канал
				messages <- delivery.Body

				// Подтверждаем обработку
				if err := delivery.Ack(false); err != nil {
					log.Printf("Failed to ack message: %v", err)
				}
			}
		}
	}()

	return messages, nil
}

func (a *RabbitMQAdapter) Close() error {
	if a.channel != nil {
		if err := a.channel.Close(); err != nil {
			log.Printf("Failed to close channel: %v", err)
		}
	}
	if a.conn != nil {
		if err := a.conn.Close(); err != nil {
			log.Printf("Failed to close connection: %v", err)
		}
	}
	return nil
}

// Implement Queue interface
var _ Queue = (*RabbitMQAdapter)(nil)
