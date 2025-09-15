package di

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pozedorum/WB_project_3/task6/internal/interfaces"
	"github.com/pozedorum/WB_project_3/task6/internal/repository"
	"github.com/pozedorum/WB_project_3/task6/internal/server"
	"github.com/pozedorum/WB_project_3/task6/internal/service"
	"github.com/pozedorum/WB_project_3/task6/pkg/config"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/ginext"
)

type Container struct {
	DB         *dbpg.DB
	HTTPServer *http.Server

	SaleRepo      interfaces.SaleRepository
	AnalyticsRepo interfaces.AnalyticsRepository
	SaleService   interfaces.SaleService
	SaleServer    interfaces.SaleServer

	closers []interfaces.Closer //Список ресурсов, которые нужно закрыть
}

func NewContainer(cfg config.Config) (*Container, error) {
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
	container.DB = db

	// Инициализация репозиториев с ОДНИМ подключением
	container.SaleRepo = repository.NewSalesTrackerRepository(db)
	container.AnalyticsRepo = repository.NewAnalyticsTrackerRepository(db)

	// Инициализация сервисов
	container.SaleService = service.New(
		container.SaleRepo,
		container.AnalyticsRepo,
	)

	// Инициализация сервера
	container.SaleServer = server.New(container.SaleService)

	// Настройка HTTP сервера
	router := ginext.New()
	apiRouter := router.Group("")
	container.SaleServer.SetupRoutes(router, apiRouter)

	serverAddr := ":" + cfg.Server.Port
	container.HTTPServer = &http.Server{
		Addr:         serverAddr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &container, nil
}

func (c *Container) Start() error {
	return c.HTTPServer.ListenAndServe()
}

func (c *Container) Shutdown(ctx context.Context) error {
	var errors []error
	// Закрываем сначала HTTP сервер, потом БД
	if c.HTTPServer != nil {
		if err := c.HTTPServer.Shutdown(ctx); err != nil {
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
	return nil
}
