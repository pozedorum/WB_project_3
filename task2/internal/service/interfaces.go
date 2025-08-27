package service

import (
	"context"

	"github.com/pozedorum/WB_project_3/task2/internal/models"
)

// Repository интерфейс для работы с данными
type Repository interface {
	CreateShortURL(ctx context.Context, n *models.ShortURL) error
	GetOriginalURLIfExists(ctx context.Context, shortCode string) (*models.ShortURL, error)
	GetStatisticsByShortCode(ctx context.Context, shortCode string, period string, groupBy string) (*models.AnalyticsResponse, error)
	RegisterClick(ctx context.Context, click *models.ClickAnalyticsEntry) error
}
