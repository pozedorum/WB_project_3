package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/pozedorum/wbf/zlog"
)

type Config struct {
	Server  ServerConfig
	Storage StorageConfig
	Retry   RetryConfig
	Kafka   KafkaConfig
}

type ServerConfig struct {
	Port string
}

type StorageConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	Region          string
	UseSSL          bool
}

type RetryConfig struct {
	MaxRetries  int
	BaseDelay   time.Duration
	WorkerCount int
}

type KafkaConfig struct {
	Brokers []string
	Topic   string
	GroupID string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		zlog.Logger.Info().Msg("No .env file found, using environment variables")
	}

	useSSL, _ := strconv.ParseBool(getEnv("MINIO_USE_SSL", "false"))

	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		Storage: StorageConfig{
			Endpoint:        getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKeyID:     getEnv("MINIO_ACCESS_KEY", "minioadmin"),
			SecretAccessKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
			BucketName:      getEnv("MINIO_BUCKET", "images"),
			Region:          getEnv("MINIO_REGION", "us-east-1"),
			UseSSL:          useSSL,
		},
		Kafka: KafkaConfig{
			Brokers: getEnvAsList("KAFKA_BROKERS", []string{"localhost:9092"}), // Значение по умолчанию
			Topic:   getEnv("KAFKA_TOPIC", "image-processing-tasks"),           // Значение по умолчанию
			GroupID: getEnv("KAFKA_GROUPID", "image-processor-service"),        // Значение по умолчанию
		},
		Retry: RetryConfig{
			MaxRetries:  getEnvAsInt("MAX_RETRIES", 3),
			BaseDelay:   getEnvAsDuration("BASE_DELAY", 1*time.Second),
			WorkerCount: getEnvAsInt("WORKER_COUNT", 5),
		},
	}
}

// ValidateConfig проверяет валидность конфигурации
func (cfg *Config) ValidateConfig() error {
	// Проверка Storage
	if cfg.Storage.Endpoint == "" {
		return fmt.Errorf("MINIO_ENDPOINT is required")
	}
	if cfg.Storage.AccessKeyID == "" {
		return fmt.Errorf("MINIO_ACCESS_KEY is required")
	}
	if cfg.Storage.SecretAccessKey == "" {
		return fmt.Errorf("MINIO_SECRET_KEY is required")
	}
	if cfg.Storage.BucketName == "" {
		return fmt.Errorf("MINIO_BUCKET is required")
	}

	// Проверка Kafka
	if len(cfg.Kafka.Brokers) == 0 {
		return fmt.Errorf("KAFKA_BROKERS is required")
	}
	if cfg.Kafka.Topic == "" {
		return fmt.Errorf("KAFKA_TOPIC is required")
	}
	if cfg.Kafka.GroupID == "" {
		return fmt.Errorf("KAFKA_GROUPID is required")
	}

	return nil
}

// Вспомогательные функции
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsList(key string, defaultValue []string) []string {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}

	// Убираем пробелы и разбиваем по запятым
	values := strings.Split(valueStr, ",")
	for i, v := range values {
		values[i] = strings.TrimSpace(v)
	}
	return values
}
