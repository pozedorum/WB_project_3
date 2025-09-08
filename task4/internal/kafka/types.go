package kafka

import (
	"time"

	baseKafka "github.com/segmentio/kafka-go"
	"github.com/wb-go/wbf/kafka"
)

// KafkaConfig конфигурация Kafka
type KafkaConfig struct {
	Brokers          []string
	Topic            string
	GroupID          string
	SessionTimeout   time.Duration
	RebalanceTimeout time.Duration
	MaxWait          time.Duration
	MinBytes         int
	MaxBytes         int
}

// KafkaProducer продюсер для Kafka
type KafkaProducer struct {
	writer *kafka.Producer
	topic  string
}

// KafkaConsumer консьюмер для Kafka
type KafkaConsumer struct {
	reader *kafka.Consumer
	topic  string
}

type Message = baseKafka.Message
