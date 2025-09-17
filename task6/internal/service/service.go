package service

import (
	"context"
	"fmt"
	"time"

	"github.com/pozedorum/WB_project_3/task6/internal/interfaces"
	"github.com/pozedorum/WB_project_3/task6/internal/models"
	"github.com/shopspring/decimal"
)

type SaleTrackerService struct {
	SaleRepo      interfaces.SaleRepository
	AnalyticsRepo interfaces.AnalyticsRepository
}

func New(SaleRepo interfaces.SaleRepository, AnalyticsRepo interfaces.AnalyticsRepository) *SaleTrackerService {
	return &SaleTrackerService{SaleRepo: SaleRepo, AnalyticsRepo: AnalyticsRepo}
}

func (servs *SaleTrackerService) CreateSale(ctx context.Context, req *models.SaleRequest) (*models.SaleInformation, error) {
	// Валидация данных
	if err := servs.validateSale(req); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}
	now := time.Now()
	sale := &models.SaleInformation{
		Amount:      req.Amount,
		Type:        req.Type,
		Category:    req.Category,
		Description: req.Description,
		Date:        req.Date,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Сохранение в репозитории
	if err := servs.SaleRepo.Create(ctx, sale); err != nil {
		return nil, fmt.Errorf("failed to create sale: %w", err)
	}

	return sale, nil
}

func (servs *SaleTrackerService) GetSaleByID(ctx context.Context, id int64) (*models.SaleInformation, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid ID: must be positive integer")
	}

	sale, err := servs.SaleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get sale by ID: %w", err)
	}

	return sale, nil
}

func (servs *SaleTrackerService) GetAllSales(ctx context.Context, filters map[string]interface{}) ([]models.SaleInformation, error) {
	// Валидация и нормализация фильтров
	validatedFilters := servs.validateAndNormalizeFilters(filters)

	sales, err := servs.SaleRepo.FindAll(ctx, validatedFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to get all sales: %w", err)
	}

	return sales, nil
}

func (servs *SaleTrackerService) UpdateSale(ctx context.Context, id int64, req *models.SaleRequest) (*models.SaleInformation, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid ID: must be positive integer")
	}

	// Валидация данных
	if err := servs.validateSale(req); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Проверяем существование записи
	existingSale, err := servs.SaleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("sale not found: %w", err)
	}
	sale := &models.SaleInformation{
		ID:          id,
		Amount:      req.Amount,
		Type:        req.Type,
		Category:    req.Category,
		Description: req.Description,
		Date:        req.Date,
		CreatedAt:   existingSale.CreatedAt,
		UpdatedAt:   time.Now(),
	}

	// Обновление в репозитории
	if err := servs.SaleRepo.Update(ctx, id, sale); err != nil {
		return nil, fmt.Errorf("failed to update sale: %w", err)
	}

	return sale, nil
}

func (servs *SaleTrackerService) DeleteSale(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid ID: must be positive integer")
	}

	if err := servs.SaleRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete sale: %w", err)
	}

	return nil
}

func (servs *SaleTrackerService) GetAnalytics(ctx context.Context, req *models.AnalyticsRequest) (*models.AnalyticsResponse, error) {
	// Валидация запроса
	if err := servs.validateAnalyticsRequest(req); err != nil {
		return nil, fmt.Errorf("invalid analytics request: %w", err)
	}

	// Получение аналитики из репозитория
	analytics, err := servs.AnalyticsRepo.GetAnalytics(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get analytics: %w", err)
	}

	return analytics, nil
}

func (servs *SaleTrackerService) ExportCSV(ctx context.Context, req *models.CSVExportRequest) ([]byte, error) {
	// Валидация запроса
	if err := servs.validateCSVExportRequest(req); err != nil {
		return nil, fmt.Errorf("invalid CSV export request: %w", err)
	}

	// Если группировка не указана, экспортируем сырые данные
	if req.GroupBy == "" {
		return servs.AnalyticsRepo.ExportToCSV(ctx, req)
	}

	// Если указана группировка, получаем аналитику и преобразуем в CSV
	analyticsReq := &models.AnalyticsRequest{
		From:     req.From,
		To:       req.To,
		Type:     req.Type,
		Category: req.Category,
		GroupBy:  req.GroupBy,
	}

	analytics, err := servs.GetAnalytics(ctx, analyticsReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get analytics for CSV export: %w", err)
	}

	// Преобразование аналитики в CSV формат
	csvData, err := servs.analyticsToCSV(analytics, req.GroupBy)
	if err != nil {
		return nil, fmt.Errorf("failed to convert analytics to CSV: %w", err)
	}

	return csvData, nil
}

// Вспомогательные методы

func (servs *SaleTrackerService) validateSale(sale *models.SaleRequest) error {
	if sale.Amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("amount must be positive")
	}

	if sale.Type != "income" && sale.Type != "expense" {
		return fmt.Errorf("type must be 'income' or 'expense'")
	}

	if sale.Category == "" {
		return fmt.Errorf("category is required")
	}

	if sale.Date.IsZero() {
		return fmt.Errorf("date is required")
	}

	// Проверка, что дата не в будущем
	if sale.Date.After(time.Now()) {
		return fmt.Errorf("date cannot be in the future")
	}

	return nil
}

func (servs *SaleTrackerService) validateAnalyticsRequest(req *models.AnalyticsRequest) error {
	if req.From.IsZero() {
		return fmt.Errorf("from date is required")
	}

	if req.To.IsZero() {
		return fmt.Errorf("to date is required")
	}

	if req.From.After(req.To) {
		return fmt.Errorf("from date cannot be after to date")
	}

	if req.Type != "" && req.Type != "income" && req.Type != "expense" {
		return fmt.Errorf("type must be 'income', 'expense', or empty")
	}

	if req.GroupBy != "" && req.GroupBy != "day" && req.GroupBy != "week" && req.GroupBy != "month" && req.GroupBy != "category" {
		return fmt.Errorf("group_by must be 'day', 'week', 'month', 'category', or empty")
	}

	return nil
}

func (servs *SaleTrackerService) validateCSVExportRequest(req *models.CSVExportRequest) error {
	if req.From.IsZero() {
		return fmt.Errorf("from date is required")
	}

	if req.To.IsZero() {
		return fmt.Errorf("to date is required")
	}

	if req.From.After(req.To) {
		return fmt.Errorf("from date cannot be after to date")
	}

	if req.Type != "" && req.Type != "income" && req.Type != "expense" {
		return fmt.Errorf("type must be 'income', 'expense', or empty")
	}

	if req.GroupBy != "" && req.GroupBy != "day" && req.GroupBy != "week" && req.GroupBy != "month" && req.GroupBy != "category" {
		return fmt.Errorf("group_by must be 'day', 'week', 'month', 'category', or empty")
	}

	return nil
}

func (servs *SaleTrackerService) validateAndNormalizeFilters(filters map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Копируем и валидируем фильтры
	for key, value := range filters {
		switch key {
		case "from", "to":
			if t, ok := value.(time.Time); ok && !t.IsZero() {
				result[key] = t
			}
		case "category", "type":
			if s, ok := value.(string); ok && s != "" {
				result[key] = s
			}
		case "limit":
			if limit, ok := value.(int); ok && limit > 0 {
				result[key] = limit
			}
		}
	}

	return result
}

// TODO: полностью поменять
func (servs *SaleTrackerService) analyticsToCSV(analytics *models.AnalyticsResponse, groupBy string) ([]byte, error) {

	return []byte(nil), nil
}
