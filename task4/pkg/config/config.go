package config

import (
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
	StorageType string

	// Google Drive настройки
	GoogleDriveCredentials string
	GoogleDriveToken       string // для OAuth
	GoogleDriveFolderID    string
	GoogleDriveBaseURL     string
	GoogleDriveAuthMethod  string // "service_account" или "oauth"
}

type RetryConfig struct {
	MaxRetries  int
	BaseDelay   time.Duration
	WorkerCount int
}

// TODO: дописать .env для gDrive
func Load() *Config {
	// Загрузка .env файла
	if err := godotenv.Load(); err != nil {
		zlog.Logger.Info().Msg("No .env file found, using environment variables")
	}

	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		Storage: StorageConfig{},
		Retry: RetryConfig{
			MaxRetries:  getEnvAsInt("MAX_RETRIES", 3),
			BaseDelay:   getEnvAsDuration("BASE_DELAY", 1*time.Second),
			WorkerCount: getEnvAsInt("WORKER_COUNT", 5),
		},
	}
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
