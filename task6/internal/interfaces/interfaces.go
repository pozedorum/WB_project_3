package interfaces

import (
	"context"
	"time"

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
	GetAnalytics(ctx context.Context, req *models.AnalyticsRequest) (*models.AnalyticsResponse, error)
	ExportToCSV(ctx context.Context, from, to time.Time) ([]byte, error)
}

type AnalyticsRepository interface {
	GetSalesSummary(ctx context.Context, from, to time.Time, category, saleType string) (decimal.Decimal, decimal.Decimal, int64, error)
	GetMedian(ctx context.Context, from, to time.Time, category, saleType string) (decimal.Decimal, error)
	GetPercentile90(ctx context.Context, from, to time.Time, category, saleType string) (decimal.Decimal, error)
	GetGroupedData(ctx context.Context, from, to time.Time, groupBy, category, saleType string) ([]models.GroupedDataItem, error)
}

type SaleService interface {
	CreateSale(ctx context.Context, sale *models.SaleInformation) error
	GetSaleByID(ctx context.Context, id int64) (*models.SaleInformation, error)
	GetAllSales(ctx context.Context, filters map[string]interface{}) ([]models.SaleInformation, error)
	UpdateSale(ctx context.Context, id int64, sale *models.SaleInformation) error
	DeleteSale(ctx context.Context, id int64) error
	GetAnalytics(ctx context.Context, req *models.AnalyticsRequest) (*models.AnalyticsResponse, error)
	ExportCSV(ctx context.Context, req *models.CSVExportRequest) ([]byte, error)
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
