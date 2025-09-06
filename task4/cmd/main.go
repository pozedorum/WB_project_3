package main

import (
	"log"

	"github.com/pozedorum/WB_project_3/task4/internal/config"
	"github.com/pozedorum/WB_project_3/task4/internal/processor"
	"github.com/pozedorum/WB_project_3/task4/internal/repository"
	"github.com/pozedorum/WB_project_3/task4/internal/server"
	"github.com/pozedorum/WB_project_3/task4/internal/service"
	"github.com/pozedorum/WB_project_3/task4/internal/storage"
	"github.com/pozedorum/wbf/ginext"
)

func main() {
	// Загрузка конфигурации
	cfg := config.Load()
	if err := cfg.ValidateConfig(); err != nil {
		log.Fatal("Configuration error: ", err)
	}

	// Инициализация хранилища
	storage, err := storage.NewMinIOStorage(
		cfg.Storage.Endpoint,
		cfg.Storage.AccessKeyID,
		cfg.Storage.SecretAccessKey,
		cfg.Storage.BucketName,
		cfg.Storage.Region,
		cfg.Storage.UseSSL,
	)
	if err != nil {
		log.Fatal("Failed to initialize storage: ", err)
	}
	// Инициализация локального хранилища
	repo := repository.NewRepositoryInMemory()

	// Инициализация процессора
	processor := processor.NewDefaultProcessor()

	// Инициализация сервиса (который вы еще, видимо, будете дописывать)
	imageService := service.NewImageProcessService(repo, storage, processor, nil) // Примерно так

	// Инициализация и запуск сервера
	server := server.New(imageService)
	router := ginext.New()
	router.LoadHTMLGlob("internal/frontend/templates/*.html")
	apiGroup := router.Group("")
	server.SetupRoutes(apiGroup)

	log.Printf("Server starting on port %s", cfg.Server.Port)
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
