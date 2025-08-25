package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pozedorum/WB_project_3/task1/internal/config"
	"github.com/pozedorum/WB_project_3/task1/internal/notifier"
	queue "github.com/pozedorum/WB_project_3/task1/internal/rabbitmq"
	"github.com/pozedorum/WB_project_3/task1/internal/repository/postgres"
	"github.com/pozedorum/WB_project_3/task1/internal/repository/redis"
	"github.com/pozedorum/WB_project_3/task1/internal/server"
	"github.com/pozedorum/WB_project_3/task1/internal/service"
	"github.com/pozedorum/WB_project_3/task1/internal/worker"
	"github.com/pozedorum/wbf/dbpg"
	"github.com/pozedorum/wbf/ginext"
	"github.com/pozedorum/wbf/zlog"
)

func main() {
	// Инициализация логгера
	zlog.Init()

	// 1. Загрузка конфига
	cfg := config.Load()

	zlog.Logger.Info().Interface("config", cfg).Msg("Configuration loaded")

	opts := &dbpg.Options{
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
	}

	// 2. Подключение к БД
	pgRepo, err := postgres.NewNotificationRepositoryWithDB(cfg.Database.GetDSN(), []string{}, opts)
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("Failed to connect to PostgreSQL")
	}
	defer func() {
		if pgRepo != nil {
			pgRepo.Close()
		}
		zlog.Logger.Info().Msg("PostgreSQL connection closed")
	}()

	// 3. Подключение к Redis
	redisCache := redis.NewNotificationRepository(cfg.Redis.Host+":"+cfg.Redis.Port, cfg.Redis.Password, cfg.Redis.DB)

	// Проверка подключения к Redis
	ctx := context.Background()
	if _, err = redisCache.Ping(ctx); err != nil {
		zlog.Logger.Warn().Err(err).Msg("Redis connection warning")
	} else {
		zlog.Logger.Info().Msg("Connected to Redis")
	}

	// 4. Подключение к RabbitMQ
	rabbitMQ, err := queue.NewRabbitMQAdapter(cfg.RabbitMQ.GetURL())
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("Failed to connect to RabbitMQ")
	}
	defer func() {
		if err := rabbitMQ.Close(); err != nil {
			zlog.Logger.Error().Err(err).Msg("Error closing RabbitMQ connection")
		} else {
			zlog.Logger.Info().Msg("RabbitMQ connection closed")
		}
	}()

	// 5. Создание нотификаторов
	notifiers := []service.Notifier{}
	if enot, err := notifier.NewEmailNotifier(cfg.Email); err != nil {
		zlog.Logger.Error().Err(err).Msg("Error with creating Email notifier")
	} else {
		notifiers = append(notifiers, enot)
	}
	if enot, err := notifier.NewTelegramNotifier(cfg.Telegram); err != nil {
		zlog.Logger.Error().Err(err).Msg("Error with creating Telegram notifier")
	} else {
		notifiers = append(notifiers, enot)
	}

	zlog.Logger.Info().Int("count", len(notifiers)).Msg("Notifiers initialized")

	// 6. Создание сервиса
	notificationService := service.NewNotificationService(pgRepo, redisCache, rabbitMQ, notifiers)

	// 7. Запуск HTTP-сервера
	server := server.New(notificationService)
	router := ginext.New()
	router.LoadHTMLGlob("internal/frontend/templates/*.html")
	// Создаем группу /api для всех routes
	apiGroup := router.Group("")
	server.SetupRoutes(apiGroup)

	// Запуск HTTP сервера в горутине
	serverAddr := ":" + cfg.Server.Port
	httpServer := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}

	go func() {
		zlog.Logger.Info().Str("address", serverAddr).Msg("Starting HTTP server")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zlog.Logger.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// 8. Запуск Воркера
	worker := worker.NewWorker(notificationService)
	go func() {
		if err := worker.Start(context.Background(), "notifications_queue"); err != nil {
			zlog.Logger.Fatal().Err(err).Msg("Failed to start worker")
		}
	}()
	zlog.Logger.Info().Msg("Worker started")

	// 9. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zlog.Logger.Info().Msg("Shutting down server...")

	// Создаем контекст с таймаутом для graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Останавливаем HTTP сервер
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		zlog.Logger.Error().Err(err).Msg("HTTP server shutdown error")
	} else {
		zlog.Logger.Info().Msg("HTTP server stopped gracefully")
	}

	workerStopChan := make(chan struct{})
	go func() {
		worker.Stop()
		close(workerStopChan)
	}()

	select {
	case <-workerStopChan:
		zlog.Logger.Info().Msg("Worker stopped gracefully")
	case <-time.After(10 * time.Second):
		zlog.Logger.Info().Msg("Worker shutdown timeout")
	}

	zlog.Logger.Info().Msg("Server exited properly")
}
