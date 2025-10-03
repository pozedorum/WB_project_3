package di

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pozedorum/WB_project_3/task6/pkg/config"
	"github.com/pozedorum/WB_project_3/task7/internal/interfaces"
	"github.com/pozedorum/WB_project_3/task7/internal/repository"
	"github.com/pozedorum/WB_project_3/task7/internal/server"
	"github.com/pozedorum/WB_project_3/task7/internal/service"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/ginext"
)

type Container struct {
	db         *dbpg.DB
	httpServer *http.Server

	repo    interfaces.Repository
	service interfaces.Service
	server  interfaces.Server

	closers []interfaces.Closer // Список ресурсов, которые нужно закрыть
}

func NewContainer(cfg *config.Config) (*Container, error) {
	var container Container

	// Инициализация DB
	opts := &dbpg.Options{
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
	}
	db, err := dbpg.New(cfg.Database.GetDSN(), []string{}, opts)
	if err != nil {
		return nil, err
	}
	container.db = db

	// Инициализация репозиториев
	container.repo = repository.New(db)
	container.closers = append(container.closers, container.repo.(interfaces.Closer))
	// Инициализация сервисов
	container.service = service.New(container.repo)

	// Инициализация сервера
	container.server = server.New(container.service)

	// Настройка HTTP сервера
	router := ginext.New()
	apiRouter := router.Group("")
	container.server.SetupRoutes(router, apiRouter)

	serverAddr := ":" + cfg.Server.Port
	container.httpServer = &http.Server{
		Addr:         serverAddr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &container, nil
}

func (c *Container) Start() error {
	fmt.Printf("Server starting on http://localhost%s\n", c.httpServer.Addr)
	return c.httpServer.ListenAndServe()
}

func (c *Container) Shutdown(ctx context.Context) error {
	var errors []error
	// Закрываем сначала HTTP сервер, потом БД
	fmt.Println("Shutting down server...")
	if c.httpServer != nil {
		if err := c.httpServer.Shutdown(ctx); err != nil {
			errors = append(errors, fmt.Errorf("HTTP server shutdown failed: %w", err))
		}
	}

	for i := len(c.closers) - 1; i >= 0; i-- {
		if err := c.closers[i].Close(); err != nil {
			errors = append(errors, fmt.Errorf("resource close failed: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("shutdown completed with errors: %v", errors)
	}
	fmt.Println("Shutdown completed successfully")
	return nil
}
