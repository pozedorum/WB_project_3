package service

import (
	"context"
	"database/sql"
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
	if err := validateSaleRequest(req); err != nil {
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
		return nil, models.ErrNegativeIndex
	}

	sale, err := servs.SaleRepo.FindByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrSaleNotFound
		}
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
		return nil, models.ErrNegativeIndex
	}

	// Валидация данных
	if err := validateSaleRequest(req); err != nil {
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
		return models.ErrNegativeIndex
	}

	if err := servs.SaleRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete sale: %w", err)
	}

	return nil
}

func (servs *SaleTrackerService) GetAnalytics(ctx context.Context, req *models.AnalyticsRequest) (*models.AnalyticsResponse, error) {
	// Валидация запроса
	if err := validateAnalyticsRequest(req); err != nil {
		return nil, fmt.Errorf("invalid analytics request: %w", err)
	}

	// Получение аналитики из репозитория
	analytics, err := servs.AnalyticsRepo.GetAnalytics(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get analytics: %w", err)
	}

	// Если группировки НЕТ, нужно добрать медиану и перцентиль
	if req.GroupBy == "" {
		median, err := servs.AnalyticsRepo.GetMedian(ctx, req)
		if err != nil {
			return nil, err
		}
		analytics.Median = median

		percentile90, err := servs.AnalyticsRepo.GetPercentile90(ctx, req)
		if err != nil {
			return nil, err
		}
		analytics.Percentile90 = percentile90
	}
	return analytics, nil
}

func (servs *SaleTrackerService) ExportCSV(ctx context.Context, req *models.AnalyticsRequest) ([]byte, error) {
	if err := validateAnalyticsRequest(req); err != nil {
		return nil, fmt.Errorf("invalid CSV export request: %w", err)
	}

	// ВСЕ данные теперь получаем через AnalyticsRepo
	return servs.AnalyticsRepo.ExportToCSV(ctx, req)
}

// Вспомогательные методы

func validateSaleRequest(req *models.SaleRequest) error {
	if req.Amount.LessThanOrEqual(decimal.Zero) {
		return models.ErrNegativeAmount
	}
	if req.Type != "income" && req.Type != "expense" {
		return models.ErrInvalidType
	}
	if req.Category == "" {
		return models.ErrEmptyCategory
	}
	if req.Date.IsZero() {
		return models.ErrEmptyDate
	}
	if req.Date.After(time.Now()) {
		return models.ErrWrongDate
	}
	return nil
}

func validateAnalyticsRequest(req *models.AnalyticsRequest) error {
	if req.From.IsZero() || req.To.IsZero() {
		return models.ErrEmptyFromToDate
	}

	if req.From.After(req.To) {
		return models.ErrWrongTimeRange
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
