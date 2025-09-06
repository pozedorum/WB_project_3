package repository

import (
	"context"
	"fmt"
	"sync"

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
	rim.imagesData[metadata.ID] = metadata
	rim.mu.Unlock()
	return nil
}

func (rim *RepositoryInMemory) GetImageMetadata(ctx context.Context, imageID string) (*models.ImageMetadata, error) {
	rim.mu.RLock()
	defer rim.mu.RUnlock()
	if res, ok := rim.imagesData[imageID]; ok {
		return res, nil
	}
	return nil, fmt.Errorf("metadata with id: %s in not available", imageID)
}

func (rim *RepositoryInMemory) UpdateImageStatus(ctx context.Context, imageID, status string) error {
	rim.mu.Lock()
	defer rim.mu.Unlock()
}
