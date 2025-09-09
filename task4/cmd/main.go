package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pozedorum/WB_project_3/task4/internal/config"
	"github.com/pozedorum/WB_project_3/task4/internal/kafka"
	"github.com/pozedorum/WB_project_3/task4/internal/processor"
	"github.com/pozedorum/WB_project_3/task4/internal/repository"
	"github.com/pozedorum/WB_project_3/task4/internal/server"
	"github.com/pozedorum/WB_project_3/task4/internal/service"
	"github.com/pozedorum/WB_project_3/task4/internal/storage"
	"github.com/pozedorum/wbf/ginext"
	"github.com/pozedorum/wbf/zlog"

	wbfKafka "github.com/wb-go/wbf/kafka"
)

func main() {
	// Инициализация логгера
	zlog.Init()

	// Загрузка конфигурации
	cfg := config.Load()
	if err := cfg.ValidateConfig(); err != nil {
		zlog.Logger.Fatal().Err(err).Msg("Invalid configuration")
	}

	zlog.Logger.Info().Msg("Starting Image Processing Service...")

	// Создание контекста с graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Инициализация зависимостей
	imageProcessor := processor.NewDefaultProcessor()
	repo := repository.NewRepositoryInMemory()

	// Инициализация хранилища (MinIO)
	minioStorage, err := storage.NewMinIOStorage(
		cfg.Storage.Endpoint,
		cfg.Storage.AccessKeyID,
		cfg.Storage.SecretAccessKey,
		cfg.Storage.BucketName,
		cfg.Storage.Region,
		cfg.Storage.UseSSL,
	)
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("Failed to initialize storage")
	}

	// Инициализация Kafka продюсера
	kafkaProducer := wbfKafka.NewProducer(cfg.Kafka.Brokers, cfg.Kafka.Topic)
	defer kafkaProducer.Close()

	// Инициализация Kafka очереди
	kafkaQueue := kafka.NewKafkaImageQueue(kafkaProducer, cfg.Kafka.Topic)

	// Создание сервиса
	baseURL := fmt.Sprintf("http://localhost:%s", cfg.Server.Port)
	imageService := service.NewImageProcessService(repo, minioStorage, imageProcessor, kafkaQueue, baseURL)

	// Инициализация HTTP сервера
	imageServer := server.New(imageService)
	router := ginext.New()
	router.LoadHTMLGlob("internal/frontend/templates/*.html")
	apiGroup := router.Group("")
	imageServer.SetupRoutes(apiGroup)

	// Инициализация Kafka консьюмера (если указан GroupID)
	var kafkaConsumer *kafka.KafkaTaskConsumer
	if cfg.Kafka.GroupID != "" {
		kafkaConsumerClient := wbfKafka.NewConsumer(cfg.Kafka.Brokers, cfg.Kafka.Topic, cfg.Kafka.GroupID)
		defer kafkaConsumerClient.Close()

		kafkaConsumer = kafka.NewKafkaTaskConsumer(
			kafkaConsumerClient,
			imageProcessor,
			minioStorage,
			repo,
			cfg.Kafka.Topic,
			cfg.Kafka.GroupID,
		)

		// Запуск консьюмера в отдельной горутине
		go func() {
			zlog.Logger.Info().Msg("Starting Kafka consumer...")
			kafkaConsumer.StartConsuming(ctx, cfg.Retry.WorkerCount)
		}()
	} else {
		zlog.Logger.Warn().Msg("KAFKA_GROUPID not set, running in synchronous mode only")
	}

	// Запуск HTTP сервера
	go func() {
		zlog.Logger.Info().Str("port", cfg.Server.Port).Msg("Starting HTTP server")
		if err := router.Run(":" + cfg.Server.Port); err != nil {
			zlog.Logger.Error().Err(err).Msg("HTTP server failed")
			cancel()
		}
	}()

	// Ожидание сигналов завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case sig := <-sigChan:
		zlog.Logger.Info().Str("signal", sig.String()).Msg("Received shutdown signal")
	case <-ctx.Done():
		zlog.Logger.Info().Msg("Context cancelled, shutting down")
	}

	// Graceful shutdown
	zlog.Logger.Info().Msg("Shutting down gracefully...")

	// Даем время на завершение операций
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Закрываем соединения
	if err := kafkaQueue.Close(); err != nil {
		zlog.Logger.Error().Err(err).Msg("Failed to close Kafka queue")
	}

	// Ждем завершения
	select {
	case <-shutdownCtx.Done():
		zlog.Logger.Warn().Msg("Shutdown timeout exceeded, forcing exit")
	default:
		zlog.Logger.Info().Msg("Shutdown completed successfully")
	}
}
