package interfaces

import (
	"context"

	"github.com/pozedorum/WB_project_3/task6/internal/models"
	"github.com/shopspring/decimal"
	"github.com/wb-go/wbf/ginext"
)

type SaleRepository interface {
	Create(ctx context.Context, sale *models.SaleInformation) error
	FindByID(ctx context.Context, id int64) (*models.SaleInformation, error)
	FindAll(ctx context.Context, filters map[string]interface{}) ([]models.SaleInformation, error)
	Update(ctx context.Context, id int64, sale *models.SaleInformation) error
	Delete(ctx context.Context, id int64) error
}

type AnalyticsRepository interface {
	GetAnalytics(ctx context.Context, req *models.AnalyticsRequest) (*models.AnalyticsResponse, error)
	GetSalesSummary(ctx context.Context, req *models.AnalyticsRequest) (*models.SalesSummaryResponse, error)
	GetMedian(ctx context.Context, req *models.AnalyticsRequest) (decimal.Decimal, error)
	GetPercentile90(ctx context.Context, req *models.AnalyticsRequest) (decimal.Decimal, error)
	ExportToCSV(ctx context.Context, req *models.AnalyticsRequest) ([]byte, error)
}

type SaleService interface {
	CreateSale(ctx context.Context, sale *models.SaleRequest) (*models.SaleInformation, error)
	GetSaleByID(ctx context.Context, id int64) (*models.SaleInformation, error)
	GetAllSales(ctx context.Context, filters map[string]interface{}) ([]models.SaleInformation, error)
	UpdateSale(ctx context.Context, id int64, sale *models.SaleRequest) (*models.SaleInformation, error)
	DeleteSale(ctx context.Context, id int64) error
	GetAnalytics(ctx context.Context, req *models.AnalyticsRequest) (*models.AnalyticsResponse, error)
	ExportToCSV(ctx context.Context, req *models.AnalyticsRequest) ([]byte, error)
}

type SaleServer interface {
	CreateItem(c *ginext.Context)
	GetItems(c *ginext.Context)
	GetItemByID(c *ginext.Context)
	UpdateItem(c *ginext.Context)
	DeleteItem(c *ginext.Context)

	GetAnalytics(c *ginext.Context)
	ExportCSV(c *ginext.Context)

	SetupRoutes(router *ginext.Engine, apiRouter *ginext.RouterGroup)
	ServeFrontend(c *ginext.Context)
}

// Closer интерфейс для ресурсов, которые нужно закрывать
type Closer interface {
	Close() error
}
