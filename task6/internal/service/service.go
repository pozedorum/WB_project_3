package service

import (
	"context"

	"github.com/pozedorum/WB_project_3/task6/internal/interfaces"
	"github.com/pozedorum/WB_project_3/task6/internal/models"
)

type SaleTrackerService struct {
	SaleRepo      interfaces.SaleRepository
	AnalyticsRepo interfaces.AnalyticsRepository
}

func New(SaleRepo interfaces.SaleRepository, AnalyticsRepo interfaces.AnalyticsRepository) *SaleTrackerService {
	return &SaleTrackerService{SaleRepo: SaleRepo, AnalyticsRepo: AnalyticsRepo}
}

func (servs *SaleTrackerService) CreateSale(ctx context.Context, sale *models.SaleInformation) error {
	return nil
}
func (servs *SaleTrackerService) GetSaleByID(ctx context.Context, id int64) (*models.SaleInformation, error) {
	return nil, nil
}
func (servs *SaleTrackerService) GetAllSales(ctx context.Context, filters map[string]interface{}) ([]models.SaleInformation, error) {
	return nil, nil
}
func (servs *SaleTrackerService) UpdateSale(ctx context.Context, id int64, sale *models.SaleInformation) error {
	return nil
}
func (servs *SaleTrackerService) DeleteSale(ctx context.Context, id int64) error {
	return nil
}
func (servs *SaleTrackerService) GetAnalytics(ctx context.Context, req *models.AnalyticsRequest) (*models.AnalyticsResponse, error) {
	return nil, nil
}
func (servs *SaleTrackerService) ExportCSV(ctx context.Context, req *models.CSVExportRequest) ([]byte, error) {
	return nil, nil
}
