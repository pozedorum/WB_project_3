package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/pozedorum/WB_project_3/task5/pkg/logger"
	"github.com/pozedorum/WB_project_3/task6/internal/interfaces"
	"github.com/pozedorum/WB_project_3/task6/internal/models"
	"github.com/shopspring/decimal"
	"github.com/wb-go/wbf/zlog"
)

type SaleTrackerService struct {
	SaleRepo      interfaces.SaleRepository
	AnalyticsRepo interfaces.AnalyticsRepository
}

func New(SaleRepo interfaces.SaleRepository, AnalyticsRepo interfaces.AnalyticsRepository) *SaleTrackerService {
	logger.LogService(func() {
		zlog.Logger.Info().Msg("Initializing SaleTrackerService")
	})
	return &SaleTrackerService{SaleRepo: SaleRepo, AnalyticsRepo: AnalyticsRepo}
}

func (servs *SaleTrackerService) CreateSale(ctx context.Context, req *models.SaleRequest) (*models.SaleInformation, error) {
	logger.LogService(func() {
		zlog.Logger.Info().
			Str("description", req.Description).
			Str("amount", req.Amount.String()).
			Msg("Creating new sale")
	})

	// Валидация данных
	if err := validateSaleRequest(req); err != nil {
		logger.LogService(func() {
			zlog.Logger.Warn().
				Err(err).
				Str("description", req.Description).
				Msg("Sale request validation failed")
		})
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
		logger.LogService(func() {
			zlog.Logger.Error().
				Err(err).
				Str("description", req.Description).
				Msg("Failed to create sale in repository")
		})
		return nil, fmt.Errorf("failed to create sale: %w", err)
	}

	logger.LogService(func() {
		zlog.Logger.Info().
			Int64("id", sale.ID).
			Str("description", sale.Description).
			Msg("Sale created successfully")
	})
	return sale, nil
}

func (servs *SaleTrackerService) GetSaleByID(ctx context.Context, id int64) (*models.SaleInformation, error) {
	logger.LogService(func() {
		zlog.Logger.Info().
			Int64("id", id).
			Msg("Getting sale by ID")
	})

	if id <= 0 {
		logger.LogService(func() {
			zlog.Logger.Warn().
				Int64("id", id).
				Msg("Invalid sale ID requested")
		})
		return nil, models.ErrNegativeIndex
	}

	sale, err := servs.SaleRepo.FindByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.LogService(func() {
				zlog.Logger.Warn().
					Int64("id", id).
					Msg("Sale not found")
			})
			return nil, models.ErrSaleNotFound
		}
		logger.LogService(func() {
			zlog.Logger.Error().
				Err(err).
				Int64("id", id).
				Msg("Failed to get sale by ID from repository")
		})
		return nil, fmt.Errorf("failed to get sale by ID: %w", err)
	}

	logger.LogService(func() {
		zlog.Logger.Info().
			Int64("id", id).
			Str("description", sale.Description).
			Msg("Sale retrieved successfully")
	})
	return sale, nil
}

func (servs *SaleTrackerService) GetAllSales(ctx context.Context, filters map[string]interface{}) ([]models.SaleInformation, error) {
	logger.LogService(func() {
		zlog.Logger.Info().
			Interface("filters", filters).
			Msg("Getting all sales with filters")
	})

	// Валидация и нормализация фильтров
	validatedFilters := validateAndNormalizeFilters(filters)

	sales, err := servs.SaleRepo.FindAll(ctx, validatedFilters)
	if err != nil {
		logger.LogService(func() {
			zlog.Logger.Error().
				Err(err).
				Interface("filters", validatedFilters).
				Msg("Failed to get all sales from repository")
		})
		return nil, fmt.Errorf("failed to get all sales: %w", err)
	}

	logger.LogService(func() {
		zlog.Logger.Info().
			Int("count", len(sales)).
			Interface("filters", validatedFilters).
			Msg("Sales retrieved successfully")
	})
	return sales, nil
}

func (servs *SaleTrackerService) UpdateSale(ctx context.Context, id int64, req *models.SaleRequest) (*models.SaleInformation, error) {
	logger.LogService(func() {
		zlog.Logger.Info().
			Int64("id", id).
			Str("description", req.Description).
			Msg("Updating sale")
	})

	if id <= 0 {
		logger.LogService(func() {
			zlog.Logger.Warn().
				Int64("id", id).
				Msg("Invalid sale ID for update")
		})
		return nil, models.ErrNegativeIndex
	}

	// Валидация данных
	if err := validateSaleRequest(req); err != nil {
		logger.LogService(func() {
			zlog.Logger.Warn().
				Err(err).
				Int64("id", id).
				Str("description", req.Description).
				Msg("Sale update request validation failed")
		})
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Проверяем существование записи
	_, err := servs.SaleRepo.FindByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.LogService(func() {
				zlog.Logger.Warn().
					Int64("id", id).
					Msg("Sale not found for update")
			})
			return nil, models.ErrSaleNotFound
		}
		logger.LogService(func() {
			zlog.Logger.Error().
				Err(err).
				Int64("id", id).
				Msg("Failed to find sale for update")
		})
		return nil, fmt.Errorf("failed to find sale for update: %w", err)
	}

	// Составляем новую запись
	updatedSale := &models.SaleInformation{
		ID:          id,
		Amount:      req.Amount,
		Type:        req.Type,
		Category:    req.Category,
		Description: req.Description,
		Date:        req.Date,
	}

	// Сохраняем изменения
	if err := servs.SaleRepo.Update(ctx, id, updatedSale); err != nil {
		logger.LogService(func() {
			zlog.Logger.Error().
				Err(err).
				Int64("id", id).
				Msg("Failed to update sale in repository")
		})
		return nil, fmt.Errorf("failed to update sale: %w", err)
	}

	logger.LogService(func() {
		zlog.Logger.Info().
			Int64("id", id).
			Str("description", updatedSale.Description).
			Time("updated_at", updatedSale.UpdatedAt).
			Msg("Sale updated successfully")
	})
	return updatedSale, nil
}

func (servs *SaleTrackerService) DeleteSale(ctx context.Context, id int64) error {
	logger.LogService(func() {
		zlog.Logger.Info().
			Int64("id", id).
			Msg("Deleting sale")
	})

	if id <= 0 {
		logger.LogService(func() {
			zlog.Logger.Warn().
				Int64("id", id).
				Msg("Invalid sale ID for deletion")
		})
		return models.ErrNegativeIndex
	}

	// Проверяем существование записи
	_, err := servs.SaleRepo.FindByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.LogService(func() {
				zlog.Logger.Warn().
					Int64("id", id).
					Msg("Sale not found for deletion")
			})
			return models.ErrSaleNotFound
		}
		logger.LogService(func() {
			zlog.Logger.Error().
				Err(err).
				Int64("id", id).
				Msg("Failed to find sale for deletion")
		})
		return fmt.Errorf("failed to find sale for deletion: %w", err)
	}

	// Удаляем запись
	if err := servs.SaleRepo.Delete(ctx, id); err != nil {
		logger.LogService(func() {
			zlog.Logger.Error().
				Err(err).
				Int64("id", id).
				Msg("Failed to delete sale from repository")
		})
		return fmt.Errorf("failed to delete sale: %w", err)
	}

	logger.LogService(func() {
		zlog.Logger.Info().
			Int64("id", id).
			Msg("Sale deleted successfully")
	})
	return nil
}

func (servs *SaleTrackerService) GetSalesSummary(ctx context.Context, req *models.AnalyticsRequest) (*models.SalesSummaryResponse, error) {
	logger.LogService(func() {
		zlog.Logger.Info().
			Time("from", req.From).
			Time("to", req.To).
			Str("category", req.Category).
			Str("type", req.Type).
			Msg("Getting sales summary")
	})

	// Валидация временного диапазона
	if err := validateAnalyticsRequest(req); err != nil {
		logger.LogService(func() {
			zlog.Logger.Warn().
				Err(err).
				Msg("Invalid request for sales summary")
		})
		return nil, err
	}

	summary, err := servs.AnalyticsRepo.GetSalesSummary(ctx, req)
	if err != nil {
		logger.LogService(func() {
			zlog.Logger.Error().
				Err(err).
				Time("from", req.From).
				Time("to", req.To).
				Msg("Failed to get sales summary from repository")
		})
		return nil, fmt.Errorf("failed to get sales summary: %w", err)
	}

	logger.LogService(func() {
		zlog.Logger.Info().
			Str("total_amount", summary.SumAmount.String()).
			Int64("items_count", summary.ItemsCount).
			Str("average_amount", summary.AverageAmount.String()).
			Msg("Sales summary retrieved successfully")
	})
	return summary, nil
}

func (servs *SaleTrackerService) GetMedian(ctx context.Context, req *models.AnalyticsRequest) (decimal.Decimal, error) {
	logger.LogService(func() {
		zlog.Logger.Info().
			Time("from", req.From).
			Time("to", req.To).
			Msg("Getting median value")
	})

	if err := validateAnalyticsRequest(req); err != nil {
		logger.LogService(func() {
			zlog.Logger.Warn().
				Err(err).
				Msg("Invalid request for median calculation")
		})
		return decimal.Decimal{}, err
	}

	median, err := servs.AnalyticsRepo.GetMedian(ctx, req)
	if err != nil {
		logger.LogService(func() {
			zlog.Logger.Error().
				Err(err).
				Time("from", req.From).
				Time("to", req.To).
				Msg("Failed to get median from repository")
		})
		return decimal.Decimal{}, fmt.Errorf("failed to get median: %w", err)
	}

	logger.LogService(func() {
		zlog.Logger.Info().
			Str("median", median.String()).
			Msg("Median value retrieved successfully")
	})
	return median, nil
}

func (servs *SaleTrackerService) GetPercentile90(ctx context.Context, req *models.AnalyticsRequest) (decimal.Decimal, error) {
	logger.LogService(func() {
		zlog.Logger.Info().
			Time("from", req.From).
			Time("to", req.To).
			Msg("Getting 90th percentile value")
	})

	if err := validateAnalyticsRequest(req); err != nil {
		logger.LogService(func() {
			zlog.Logger.Warn().
				Err(err).
				Msg("Invalid request for percentile calculation")
		})
		return decimal.Decimal{}, err
	}

	percentile, err := servs.AnalyticsRepo.GetPercentile90(ctx, req)
	if err != nil {
		logger.LogService(func() {
			zlog.Logger.Error().
				Err(err).
				Msg("Failed to get 90th percentile from repository")
		})
		return decimal.Decimal{}, fmt.Errorf("failed to get 90th percentile: %w", err)
	}

	logger.LogService(func() {
		zlog.Logger.Info().
			Str("percentile_90", percentile.String()).
			Msg("90th percentile value retrieved successfully")
	})
	return percentile, nil
}

func (servs *SaleTrackerService) GetAnalytics(ctx context.Context, req *models.AnalyticsRequest) (*models.AnalyticsResponse, error) {
	logger.LogService(func() {
		zlog.Logger.Info().
			Time("from", req.From).
			Time("to", req.To).
			Str("category", req.Category).
			Str("type", req.Type).
			Str("group_by", req.GroupBy).
			Msg("Getting analytics data")
	})

	if err := validateAnalyticsRequest(req); err != nil {
		logger.LogService(func() {
			zlog.Logger.Warn().
				Err(err).
				Msg("Invalid request for analytics")
		})
		return nil, err
	}

	analytics, err := servs.AnalyticsRepo.GetAnalytics(ctx, req)
	if err != nil {
		logger.LogService(func() {
			zlog.Logger.Error().
				Err(err).
				Time("from", req.From).
				Time("to", req.To).
				Str("group_by", req.GroupBy).
				Msg("Failed to get analytics from repository")
		})
		return nil, fmt.Errorf("failed to get analytics: %w", err)
	}

	logger.LogService(func() {
		zlog.Logger.Info().
			Str("total", analytics.Total.String()).
			Int64("count", analytics.Count).
			Int("grouped_items_count", len(analytics.GroupedData)).
			Msg("Analytics data retrieved successfully")
	})
	return analytics, nil
}

func (servs *SaleTrackerService) ExportToCSV(ctx context.Context, req *models.AnalyticsRequest) ([]byte, error) {
	logger.LogService(func() {
		zlog.Logger.Info().
			Time("from", req.From).
			Time("to", req.To).
			Str("group_by", req.GroupBy).
			Msg("Exporting data to CSV")
	})

	if err := validateAnalyticsRequest(req); err != nil {
		logger.LogService(func() {
			zlog.Logger.Warn().
				Err(err).
				Time("from", req.From).
				Time("to", req.To).
				Msg("Invalid date range for CSV export")
		})
		return nil, err
	}

	csvData, err := servs.AnalyticsRepo.ExportToCSV(ctx, req)
	if err != nil {
		logger.LogService(func() {
			zlog.Logger.Error().
				Err(err).
				Time("from", req.From).
				Time("to", req.To).
				Str("group_by", req.GroupBy).
				Msg("Failed to export data to CSV from repository")
		})
		return nil, fmt.Errorf("failed to export data to CSV: %w", err)
	}

	logger.LogService(func() {
		zlog.Logger.Info().
			Int("csv_size_bytes", len(csvData)).
			Msg("Data exported to CSV successfully")
	})
	return csvData, nil
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
		return models.ErrFutureDate
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

func validateAndNormalizeFilters(filters map[string]interface{}) map[string]interface{} {
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
