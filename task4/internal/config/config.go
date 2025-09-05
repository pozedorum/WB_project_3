package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/pozedorum/wbf/zlog"
)

type Config struct {
	Server  ServerConfig
	Storage StorageConfig

	Retry RetryConfig
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

func Load() *Config {
	// Загрузка .env файла
	if err := godotenv.Load(); err != nil {
		zlog.Logger.Info().Msg("No .env file found, using environment variables")
	}
	useSSL, _ := strconv.ParseBool(getEnv("MINIO_USE_SSL", "false"))
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		Storage: StorageConfig{
			Endpoint:        getEnv("MINIO_ENDPOINT", ""),
			AccessKeyID:     getEnv("MINIO_ACCESS_KEY", ""),
			SecretAccessKey: getEnv("MINIO_SECRET_KEY", ""),
			BucketName:      getEnv("MINIO_BUCKET", ""),
			Region:          getEnv("MINIO_REGION", ""),
			UseSSL:          useSSL,
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
	if cfg.Storage.Region == "" {
		return fmt.Errorf("MINIO_REGION is required")
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
