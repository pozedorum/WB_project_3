package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pozedorum/WB_project_3/task1/internal/models"
	"github.com/pozedorum/WB_project_3/task1/internal/service"
	"github.com/pozedorum/wbf/rabbitmq"
	"github.com/pozedorum/wbf/zlog"
)

type RabbitMQAdapter struct {
	conn         *rabbitmq.Connection
	channel      *rabbitmq.Channel
	mainExchange *rabbitmq.Exchange
	dlxExchange  *rabbitmq.Exchange
	publisher    *rabbitmq.Publisher
	mainQueue    rabbitmq.Queue
	delayedQueue rabbitmq.Queue
}

func NewRabbitMQAdapter(url string) (*RabbitMQAdapter, error) {
	zlog.Logger.Info().Msg("Initializing RabbitMQ adapter...")

	// Устанавливаем соединение с RabbitMQ
	conn, err := rabbitmq.Connect(url, 5, 2*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	zlog.Logger.Info().Msg("Connected to RabbitMQ")

	// Создаем канал
	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}
	zlog.Logger.Info().Msg("Channel opened")

	// Создаем основной exchange для уведомлений
	mainExchange := rabbitmq.NewExchange("notifications_exchange", "direct")
	mainExchange.Durable = true

	// Создаем Dead Letter Exchange для отложенных сообщений
	dlxExchange := rabbitmq.NewExchange("notifications_dlx", "direct")
	dlxExchange.Durable = true

	// Объявляем оба exchange
	zlog.Logger.Info().Msg("Declaring exchanges...")
	if err := mainExchange.BindToChannel(channel); err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare main exchange: %w", err)
	}
	zlog.Logger.Info().Str("exchange", mainExchange.Name()).Msg("Main exchange declared")

	if err := dlxExchange.BindToChannel(channel); err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare DLX exchange: %w", err)
	}
	zlog.Logger.Info().Str("exchange", dlxExchange.Name()).Msg("DLX exchange declared")

	// Создаем менеджер очередей
	qm := rabbitmq.NewQueueManager(channel)

	// Создаем основную очередь для уведомлений
	zlog.Logger.Info().Msg("Declaring main queue...")
	mainQueue, err := qm.DeclareQueue("notifications_queue", rabbitmq.QueueConfig{
		Durable:    true,
		AutoDelete: false,
	})
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare main queue: %w", err)
	}
	zlog.Logger.Info().Str("queue", mainQueue.Name).Msg("Main queue declared")

	// Создаем очередь для отложенных сообщений
	zlog.Logger.Info().Msg("Declaring delayed queue with DLX settings...")
	delayedQueue, err := qm.DeclareQueue("notifications_delayed_queue", rabbitmq.QueueConfig{
		Durable:    true,
		AutoDelete: false,
		Args: map[string]interface{}{
			"x-dead-letter-exchange":    mainExchange.Name(),
			"x-dead-letter-routing-key": "notifications_routing_key",
			"x-message-ttl":             int32(86400000),
		},
	})
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare delayed queue: %w", err)
	}
	zlog.Logger.Info().Str("queue", delayedQueue.Name).Str("dlx", mainExchange.Name()).Msg("Delayed queue declared")

	// Привязываем основную очередь к основному exchange
	zlog.Logger.Info().Msg("Binding main queue to main exchange...")
	err = channel.QueueBind(
		mainQueue.Name,
		"notifications_routing_key",
		mainExchange.Name(),
		false,
		nil,
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to bind main queue: %w", err)
	}
	zlog.Logger.Info().Str("routing_key", "notifications_routing_key").Msg("Main queue bound to exchange")

	// Привязываем отложенную очередь к DLX
	zlog.Logger.Info().Msg("Binding delayed queue to DLX exchange...")
	err = channel.QueueBind(
		delayedQueue.Name,
		"delayed_routing_key",
		dlxExchange.Name(),
		false,
		nil,
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to bind delayed queue to DLX: %w", err)
	}
	zlog.Logger.Info().Str("routing_key", "delayed_routing_key").Msg("Delayed queue bound to DLX")

	// Создаем publisher для отправки сообщений в DLX
	publisher := rabbitmq.NewPublisher(channel, dlxExchange.Name())
	zlog.Logger.Info().Msg("Publisher created for DLX exchange")

	zlog.Logger.Info().Msg("RabbitMQ adapter initialized successfully")
	return &RabbitMQAdapter{
		conn:         conn,
		channel:      channel,
		mainExchange: mainExchange,
		dlxExchange:  dlxExchange,
		publisher:    publisher,
		mainQueue:    mainQueue,
		delayedQueue: delayedQueue,
	}, nil
}

func (a *RabbitMQAdapter) PublishWithDelay(ctx context.Context, queueName string, message interface{}, delay time.Duration) error {
	// Сериализуем сообщение в JSON
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Вычисляем TTL в миллисекундах
	ttlMs := delay.Milliseconds()
	if ttlMs < 0 {
		ttlMs = 0
	}

	zlog.Logger.Info().Int64("ttl_ms", ttlMs).Dur("delay", delay).Msg("Publishing message to DLX")

	// Публикуем сообщение в DLX с TTL
	err = a.publisher.PublishWithRetry(
		data,
		"delayed_routing_key",
		"application/json",
		models.StandartStrategy,
		rabbitmq.PublishingOptions{
			Expiration: time.Duration(ttlMs) * time.Millisecond,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message with delay: %w", err)
	}

	zlog.Logger.Info().Dur("delay", delay).Int64("ttl_ms", ttlMs).Msg("Published message with delay")
	return nil
}

func (a *RabbitMQAdapter) Consume(ctx context.Context, queueName string) (<-chan []byte, error) {
	zlog.Logger.Info().Str("queue", a.mainQueue.Name).Msg("Starting consumer")

	// Создаем канал для сообщений
	messages := make(chan []byte, 100)

	// Создаем конфигурацию потребителя
	consumerConfig := &rabbitmq.ConsumerConfig{
		Queue:   a.mainQueue.Name,
		AutoAck: false,
	}

	// Создаем потребителя
	consumer := rabbitmq.NewConsumer(a.channel, consumerConfig)

	// Запускаем потребление в отдельной горутине
	go func() {
		defer close(messages)
		zlog.Logger.Info().Str("queue", a.mainQueue.Name).Msg("Consumer goroutine started")

		// Создаем канал для получения сообщений от RabbitMQ
		rabbitMsgs := make(chan []byte, 100)

		// Запускаем потребитель с retry стратегией
		go func() {
			zlog.Logger.Info().Str("queue", a.mainQueue.Name).Msg("Starting RabbitMQ consumer")
			err := consumer.ConsumeWithRetry(
				rabbitMsgs,
				models.ConsumerStrategy,
			)
			if err != nil {
				zlog.Logger.Error().Err(err).Msg("Consumer error")
				close(rabbitMsgs)
			}
		}()

		// Пересылаем сообщения в выходной канал
		for {
			select {
			case <-ctx.Done():
				zlog.Logger.Info().Msg("Consumer stopped by context")
				return

			case msg, ok := <-rabbitMsgs:
				if !ok {
					zlog.Logger.Info().Msg("RabbitMQ messages channel closed")
					return
				}

				zlog.Logger.Info().Str("queue", a.mainQueue.Name).Int("size", len(msg)).Msg("Received message from RabbitMQ")

				var processedMsg []byte
				if len(msg) > 0 && msg[0] == '"' && msg[len(msg)-1] == '"' {
					processedMsg = msg[1 : len(msg)-1]
					zlog.Logger.Debug().Str("message", string(processedMsg)).Msg("Unquoted message")
				} else {
					processedMsg = msg
					zlog.Logger.Debug().Str("message", string(processedMsg)).Msg("Raw message")
				}

				// Передаем сообщение дальше
				select {
				case messages <- processedMsg:
					zlog.Logger.Debug().Msg("Message forwarded to worker channel")
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	zlog.Logger.Info().Str("queue", a.mainQueue.Name).Msg("Consumer successfully started")
	return messages, nil
}

func (a *RabbitMQAdapter) Close() error {
	var errors []error
	zlog.Logger.Info().Msg("Closing RabbitMQ adapter...")

	if a.channel != nil {
		if err := a.channel.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close channel: %w", err))
		} else {
			zlog.Logger.Info().Msg("Channel closed")
		}
	}

	if a.conn != nil {
		if err := a.conn.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close connection: %w", err))
		} else {
			zlog.Logger.Info().Msg("Connection closed")
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors during close: %v", errors)
	}

	zlog.Logger.Info().Msg("RabbitMQ connection closed successfully")
	return nil
}

// Implement Queue interface
var _ service.Queue = (*RabbitMQAdapter)(nil)
