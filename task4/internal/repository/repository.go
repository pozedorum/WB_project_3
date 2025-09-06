package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pozedorum/WB_project_3/task4/internal/models"
)

type RepositoryInMemory struct {
	mu         sync.RWMutex
	imagesData map[string]*models.ImageMetadata
}

func NewRepositoryInMemory() *RepositoryInMemory {
	return &RepositoryInMemory{
		imagesData: make(map[string]*models.ImageMetadata),
	}
}

func (rim *RepositoryInMemory) SaveImageMetadata(ctx context.Context, metadata *models.ImageMetadata) error {
	rim.mu.Lock()
	defer rim.mu.Unlock()
	rim.imagesData[metadata.ID] = metadata

	return nil
}

func (rim *RepositoryInMemory) GetImageMetadata(ctx context.Context, imageID string) (*models.ImageMetadata, error) {
	rim.mu.RLock()
	defer rim.mu.RUnlock()

	metadata, exists := rim.imagesData[imageID]
	if !exists {
		return nil, fmt.Errorf("image not found: %s", imageID)
	}
	return metadata, nil
}

func (rim *RepositoryInMemory) UpdateImageStatus(ctx context.Context, imageID, status string) error {
	rim.mu.Lock()
	defer rim.mu.Unlock()

	metadata, exists := rim.imagesData[imageID]
	if !exists {
		return fmt.Errorf("image not found: %s", imageID)
	}

	metadata.Status = status
	if status == "completed" || status == "failed" {
		metadata.ProcessedAt = time.Now()
	}
	return nil
}

func (rim *RepositoryInMemory) DeleteImageMetadata(ctx context.Context, imageID string) error {
	rim.mu.Lock()
	defer rim.mu.Unlock()

	if _, exists := rim.imagesData[imageID]; !exists {
		return fmt.Errorf("image not found: %s", imageID)

	}
	delete(rim.imagesData, imageID)
	return nil
}
