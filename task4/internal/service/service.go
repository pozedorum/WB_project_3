package service

import (
	"context"
	"fmt"
	"time"

	"github.com/pozedorum/WB_project_3/task4/internal/models"
	"github.com/pozedorum/WB_project_3/task4/internal/storage"
)

type ImageProcessService struct {
	repo      Repository
	storage   Storage
	processor ImageProcessor
	queue     Queue
}

// NewImageProcessService создает новый сервис обработки изображений
func NewImageProcessService(repo Repository, storage Storage, processor ImageProcessor, queue Queue) *ImageProcessService {
	return &ImageProcessService{
		repo:      repo,
		storage:   storage,
		processor: processor,
		queue:     queue,
	}
}

// UploadImage загружает изображение и запускает обработку
func (s *ImageProcessService) UploadImage(ctx context.Context, imageData []byte, filename string, opts models.ProcessingOptions) (*models.UploadResult, error) {
	// Генерируем ID для изображения
	imageID := storage.GenerateFilename(filename)

	// Сохраняем оригинал в хранилище
	fileInfo, err := s.storage.Save(ctx, imageData, "originals/"+imageID)
	if err != nil {
		return nil, fmt.Errorf("failed to save original image: %w", err)
	}

	// Сохраняем метаданные
	metadata := &models.ImageMetadata{
		ID:           imageID,
		OriginalName: filename,
		FileName:     "originals/" + imageID,
		Status:       "uploaded",
		UploadedAt:   time.Now(),
		Width:        0, // Заполнится после обработки
		Height:       0,
		Size:         fileInfo.Size,
		Format:       fileInfo.MimeType,
		Options:      opts,
	}

	if err := s.repo.SaveImageMetadata(ctx, metadata); err != nil {
		// Пытаемся удалить сохраненный файл при ошибке
		s.storage.Delete(ctx, "originals/"+imageID)
		return nil, fmt.Errorf("failed to save metadata: %w", err)
	}

	// Создаем задачу на обработку
	task := &models.ProcessingTask{
		TaskID:    imageID,
		ImageData: imageData,
		Options:   opts,
	}

	// Публикуем задачу в очередь
	if err := s.queue.PublishTask(ctx, task); err != nil {
		s.repo.UpdateImageStatus(ctx, imageID, "failed")
		return nil, fmt.Errorf("failed to publish task: %w", err)
	}

	// Обновляем статус
	s.repo.UpdateImageStatus(ctx, imageID, "processing")

	return &models.UploadResult{
		ImageID: imageID,
		Status:  "processing",
		Message: "Image uploaded and queued for processing",
	}, nil
}

// GetImage возвращает обработанное изображение
func (s *ImageProcessService) GetImage(ctx context.Context, imageID string) (*models.ImageResult, error) {
	// Получаем метаданные
	metadata, err := s.repo.GetImageMetadata(ctx, imageID)
	if err != nil {
		return nil, fmt.Errorf("image not found: %w", err)
	}

	// Если изображение еще обрабатывается
	if metadata.Status != "completed" {
		return &models.ImageResult{
			Metadata: metadata,
		}, nil
	}

	// Получаем обработанное изображение
	processedData, err := s.storage.Get(ctx, "processed/"+imageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get processed image: %w", err)
	}

	return &models.ImageResult{
		ImageData: processedData,
		Metadata:  metadata,
	}, nil
}

// GetImageStatus возвращает статус обработки изображения
func (s *ImageProcessService) GetImageStatus(ctx context.Context, imageID string) (string, error) {
	metadata, err := s.repo.GetImageMetadata(ctx, imageID)
	if err != nil {
		return "", fmt.Errorf("image not found: %w", err)
	}
	return metadata.Status, nil
}

// DeleteImage удаляет изображение и его метаданные
func (s *ImageProcessService) DeleteImage(ctx context.Context, imageID string) error {
	// Удаляем оригинал
	if err := s.storage.Delete(ctx, "originals/"+imageID); err != nil {
		return fmt.Errorf("failed to delete original image: %w", err)
	}

	// Удаляем обработанную версию (если существует)
	if exists, _ := s.storage.Exists(ctx, "processed/"+imageID); exists {
		if err := s.storage.Delete(ctx, "processed/"+imageID); err != nil {
			return fmt.Errorf("failed to delete processed image: %w", err)
		}
	}

	// Удаляем метаданные
	if err := s.repo.DeleteImageMetadata(ctx, imageID); err != nil {
		return fmt.Errorf("failed to delete metadata: %w", err)
	}

	return nil
}

// ProcessImageSync синхронная обработка изображения (для тестирования без очереди)
func (s *ImageProcessService) ProcessImageSync(ctx context.Context, imageID string) error {
	// Получаем метаданные
	metadata, err := s.repo.GetImageMetadata(ctx, imageID)
	if err != nil {
		return fmt.Errorf("image not found: %w", err)
	}

	// Получаем оригинальное изображение
	imageData, err := s.storage.Get(ctx, metadata.FileName)
	if err != nil {
		return fmt.Errorf("failed to get original image: %w", err)
	}

	// Обрабатываем изображение
	start := time.Now()
	result, err := s.processor.ProcessImage(imageData, models.ProcessingOptions(metadata.Options))
	if err != nil {
		s.repo.UpdateImageStatus(ctx, imageID, "failed")
		return fmt.Errorf("image processing failed: %w", err)
	}

	// Сохраняем обработанное изображение
	_, err = s.storage.Save(ctx, result.ProcessedData, "processed/"+imageID)
	if err != nil {
		s.repo.UpdateImageStatus(ctx, imageID, "failed")
		return fmt.Errorf("failed to save processed image: %w", err)
	}

	// Обновляем метаданные
	metadata.Status = "completed"
	metadata.ProcessedAt = time.Now()
	metadata.ProcessingTime = result.ProcessingTime
	metadata.Width = result.Width
	metadata.Height = result.Height
	metadata.Size = result.Size
	metadata.Format = result.Format
	metadata.ProcessedName = "processed/" + imageID

	if err := s.repo.SaveImageMetadata(ctx, metadata); err != nil {
		return fmt.Errorf("failed to update metadata: %w", err)
	}

	return nil
}
