package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pozedorum/WB_project_3/task3/internal/config"
	"github.com/pozedorum/WB_project_3/task3/internal/repository/postgres"
	"github.com/pozedorum/WB_project_3/task3/internal/server"
	"github.com/pozedorum/WB_project_3/task3/internal/service"
	"github.com/pozedorum/wbf/dbpg"
	"github.com/pozedorum/wbf/ginext"
	"github.com/pozedorum/wbf/zlog"
)

func main() {
	zlog.Init()

	cfg := config.Load()

	zlog.Logger.Info().Interface("config", cfg).Msg("Configuration loaded")

	opts := &dbpg.Options{
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
	}

	pgRepo, err := postgres.NewCommentRepositoryWithDB(cfg.Database.GetDSN(), []string{}, opts)
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("Failed to connect to PostgreSQL")
	}
	defer func() {
		if pgRepo != nil {
			pgRepo.Close()
		}
		zlog.Logger.Info().Msg("PostgreSQL connection closed")
	}()

	commentService := service.NewCommentService(pgRepo)
	server := server.New(commentService)
	router := ginext.New()
	router.LoadHTMLGlob("internal/frontend/templates/*.html")
	apiGroup := router.Group("")
	server.SetupRoutes(apiGroup)

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

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zlog.Logger.Info().Msg("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Останавливаем HTTP сервер
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		zlog.Logger.Error().Err(err).Msg("HTTP server shutdown error")
	} else {
		zlog.Logger.Info().Msg("HTTP server stopped gracefully")
	}
}
