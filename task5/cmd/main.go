package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pozedorum/WB_project_3/task5/internal/repository"
	"github.com/pozedorum/WB_project_3/task5/internal/server"
	"github.com/pozedorum/WB_project_3/task5/internal/service"
	"github.com/pozedorum/WB_project_3/task5/pkg/config"
	"github.com/pozedorum/WB_project_3/task5/pkg/logger"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	// Инициализация логгера
	zlog.Init()
	logger.Init(true, true, true)
	// Загрузка конфигурации
	cfg := config.Load()

	zlog.Logger.Info().Msg("Starting Event Booking Service...")
	// Инициализация базы данных
	opts := &dbpg.Options{
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
	}
	pgRepo, err := repository.NewEventBookerRepositoryWithDB(cfg.Database.GetDSN(), []string{}, opts)
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("Failed to connect to PostgreSQL")
	}
	defer func() {
		if pgRepo != nil {
			pgRepo.Close()
		}
		zlog.Logger.Info().Msg("PostgreSQL connection closed")
	}()
	// Инициализация сервиса

	eventBookerService := service.NewEventBookerService(pgRepo)

	// Инициализация сервера
	eventBookerServer := server.New(eventBookerService, &cfg.JWT)
	// Запуск фонового worker для обработки просроченных бронирований
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	eventBookerService.StartCronWorker(ctx)
	// Создание роутера
	router := ginext.New()
	apiRouter := router.Group("")
	router.LoadHTMLGlob("internal/frontend/templates/*.html")
	// Настройка маршрутов
	eventBookerServer.SetupRoutes(apiRouter)
	serverAddr := ":" + cfg.Server.Port
	httpServer := &http.Server{
		Addr:         serverAddr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Запуск HTTP сервера
	go func() {
		zlog.Logger.Info().Str("address", serverAddr).Msg("Starting HTTP server")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zlog.Logger.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zlog.Logger.Info().Msg("Shutting down server...")
	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Останавливаем HTTP сервер
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		zlog.Logger.Error().Err(err).Msg("HTTP server shutdown error")
	} else {
		zlog.Logger.Info().Msg("HTTP server stopped gracefully")
	}
	// Даем время для завершения работы cron worker
	time.Sleep(time.Second)
}
