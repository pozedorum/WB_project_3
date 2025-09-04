package service

import (
	"context"
	"image"

	"github.com/pozedorum/WB_project_3/task4/pkg/processor"
	"github.com/pozedorum/WB_project_3/task4/pkg/storage"
)

type Repository interface {
}

type ImageProcessor interface {
	ProcessImage(imageData []byte, options processor.ProcessingOptions) (*processor.ProcessingResult, error)
	Resize(img image.Image, width, height int) (image.Image, error)
	AddWatermark(img image.Image, text string) (image.Image, error)
	CreateThumbnail(img image.Image, size int) (image.Image, error)
	ConvertFormat(img image.Image, format string, quality int) ([]byte, error)
}

type Storage interface {
	Save(ctx context.Context, data []byte, filename string) (*storage.FileInfo, error)
	Get(ctx context.Context, filename string) ([]byte, error)
	Delete(ctx context.Context, filename string) error
	Exists(ctx context.Context, filename string) (bool, error)
	GetURL(ctx context.Context, filename string) (string, error) // Для HTTP доступа
}
