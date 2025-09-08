package service

import (
	"context"
	"image"

	"github.com/pozedorum/WB_project_3/task4/internal/models"
)

type ImageProcessor interface {
	ProcessImage(imageData []byte, options models.ProcessingOptions) (*models.ProcessingResult, error)
	Resize(img image.Image, width, height int) (image.Image, error)
	AddWatermark(img image.Image, text string) (image.Image, error)
	CreateThumbnail(img image.Image, size int) (image.Image, error)
	ConvertFormat(img image.Image, format string, quality int) ([]byte, error)
}

type Storage interface {
	// Основные операции
	Save(ctx context.Context, data []byte, filename string) (*models.FileInfo, error)
	Get(ctx context.Context, filename string) ([]byte, error)
	Delete(ctx context.Context, filename string) error
	Exists(ctx context.Context, filename string) (bool, error)
	GetURL(ctx context.Context, filename string) (string, error)

	// Дополнительные операции
	GetFileInfo(ctx context.Context, filename string) (*models.FileInfo, error)
	ListFiles(ctx context.Context, prefix string) ([]models.FileInfo, error)
}

// Queue заглушка для очереди
// ImageQueue интерфейс для очереди обработки изображений
type ImageQueue interface {
	PublishImageTask(ctx context.Context, task *models.ProcessingTask) error
	HealthCheck(ctx context.Context) error
	Close() error
}

// Repository для работы с метаданными изображений
type Repository interface {
	SaveImageMetadata(ctx context.Context, metadata *models.ImageMetadata) error
	GetImageMetadata(ctx context.Context, imageID string) (*models.ImageMetadata, error)
	UpdateImageStatus(ctx context.Context, imageID, status string) error
	DeleteImageMetadata(ctx context.Context, imageID string) error
}
