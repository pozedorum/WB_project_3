package repository

import (
	"context"
	"database/sql" // нужен только для sql.Row
	"fmt"
	"strings"
	"time"

	"github.com/pozedorum/WB_project_3/task5/pkg/logger"
	"github.com/pozedorum/WB_project_3/task6/internal/models"
	"github.com/shopspring/decimal"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"
)

type AnalyticsTrackerRepository struct {
	db *dbpg.DB
}

func NewAnalyticsTrackerRepositoryWithDB(masterDSN string, slaveDSNs []string, opts *dbpg.Options) (*AnalyticsTrackerRepository, error) {
	db, err := dbpg.New(masterDSN, slaveDSNs, opts)
	if err != nil {
		return nil, err
	}
	return NewAnalyticsTrackerRepository(db), nil
}

func NewAnalyticsTrackerRepository(db *dbpg.DB) *AnalyticsTrackerRepository {
	return &AnalyticsTrackerRepository{db: db}
}

func (repo *AnalyticsTrackerRepository) Close() {
	if err := repo.db.Master.Close(); err != nil {
		logger.LogRepository(func() { zlog.Logger.Panic().Msg("Database failed to close") })
	}
	for _, slave := range repo.db.Slaves {
		if slave != nil {
			if err := slave.Close(); err != nil {
				logger.LogRepository(func() { zlog.Logger.Panic().Msg("Slave database failed to close") })
			}
		}
	}
	logger.LogRepository(func() { zlog.Logger.Info().Msg("PostgreSQL connections closed") })
}

func (repo *AnalyticsTrackerRepository) GetSalesSummary(ctx context.Context, req *models.AnalyticsRequest) (*models.SalesSummaryResponse, error) {
	query := `SELECT SUM(amount), COUNT(*), AVG(amount)`
	var result models.SalesSummaryResponse

	err := retry.Do(func() error {
		return repo.buildAndExecuteEasyQuery(ctx, query, req.From, req.To, req.Category, req.Type).
			Scan(&result.SumAmount, &result.ItemsCount, &result.AverageAmount)
	}, models.StandartStrategy)
	if err != nil {
		logger.LogRepository(func() {
			zlog.Logger.Error().Err(err).Msg("Error in SalesSummary query")
		})
		return &models.SalesSummaryResponse{}, err
	}
	return &result, nil
}

func (repo *AnalyticsTrackerRepository) GetMedian(ctx context.Context, req *models.AnalyticsRequest) (decimal.Decimal, error) {
	query := `SELECT PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY amount)`

	var result decimal.Decimal

	err := retry.Do(func() error {
		return repo.buildAndExecuteEasyQuery(ctx, query, req.From, req.To, req.Category, req.Type).
			Scan(&result)
	}, models.StandartStrategy)

	if err != nil {
		logger.LogRepository(func() {
			zlog.Logger.Error().Err(err).Msg("Error in GetMedian query")
		})
		return decimal.Decimal{}, err
	}
	return result, nil
}

func (repo *AnalyticsTrackerRepository) GetPercentile90(ctx context.Context, req *models.AnalyticsRequest) (decimal.Decimal, error) {
	query := `SELECT PERCENTILE_CONT(0.9) WITHIN GROUP (ORDER BY amount)`

	var result decimal.Decimal

	err := retry.Do(func() error {
		return repo.buildAndExecuteEasyQuery(ctx, query, req.From, req.To, req.Category, req.Type).
			Scan(&result)
	}, models.StandartStrategy)

	if err != nil {
		logger.LogRepository(func() {
			zlog.Logger.Error().Err(err).Msg("Error in GetPercentile90 query")
		})
		return decimal.Decimal{}, err
	}
	return result, nil
}

func (repo *AnalyticsTrackerRepository) GetAnalytics(ctx context.Context, req *models.AnalyticsRequest) (*models.AnalyticsResponse, error) {
	// 1. Строим запрос
	query, args, err := buildAnalyticsQuery(req)
	if err != nil {
		return nil, err
	}

	// 2. Выполняем запрос
	response, err := repo.executeAnalyticsQuery(ctx, query, args, req)
	if err != nil {
		logger.LogRepository(func() {
			zlog.Logger.Error().Err(err).Msg("Error in GetAnalytics query")
		})
		return nil, err
	}

	return response, nil
}

func (repo *AnalyticsTrackerRepository) ExportToCSV(ctx context.Context, req *models.AnalyticsRequest) ([]byte, error) {
	return nil, nil
}

func (repo *AnalyticsTrackerRepository) executeAnalyticsQuery(ctx context.Context, query string, args []interface{}, req *models.AnalyticsRequest) (*models.AnalyticsResponse, error) {
	var response models.AnalyticsResponse

	err := retry.Do(func() error {
		if req.GroupBy == "" {
			// Запрос БЕЗ группировки - только AnalyticsResponse
			return repo.db.Master.QueryRowContext(ctx, query, args...).Scan(
				&response.Total,
				&response.Average,
				&response.Count,
				&response.Median,
				&response.Percentile90,
			)
		} else {
			// Запрос С группировкой - AnalyticsResponse и заполненное поле GroupedDataItem
			return repo.processGroupedQuery(ctx, query, args, &response)
		}
	}, models.StandartStrategy)

	return &response, err
}

func (repo *AnalyticsTrackerRepository) buildAndExecuteEasyQuery(ctx context.Context, firstPartQuery string, from, to time.Time, category, saleType string) *sql.Row {
	var bldr strings.Builder
	bldr.WriteString(firstPartQuery)
	bldr.WriteString(` FROM sales
		WHERE date BETWEEN $1 AND $2`)

	args := []interface{}{from, to}
	argIndex := 3

	// Добавляем фильтры
	if category != "" {
		bldr.WriteString(fmt.Sprintf(" AND category = $%d", argIndex))
		args = append(args, category)
		argIndex++
	}

	if saleType != "" {
		bldr.WriteString(fmt.Sprintf(" AND type = $%d", argIndex))
		args = append(args, saleType)
		argIndex++
	}
	return repo.db.Master.QueryRowContext(ctx, bldr.String(), args...)

}

func (repo *AnalyticsTrackerRepository) processGroupedQuery(ctx context.Context, query string, args []interface{}, response *models.AnalyticsResponse) error {
	rows, err := repo.db.Master.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	var groupedData []models.GroupedDataItem
	var grandTotal decimal.Decimal
	var totalCount int64

	for rows.Next() {
		var item models.GroupedDataItem
		err := rows.Scan(
			&item.Group,
			&item.Total,
			&item.Average,
			&item.Count,
			&item.Median,
			&item.Percentile90,
			&item.Min,
			&item.Max,
		)
		if err != nil {
			return err
		}

		groupedData = append(groupedData, item)
		grandTotal = grandTotal.Add(item.Total)
		totalCount += item.Count
	}

	// Заполняем общую статистику
	response.GroupedData = groupedData
	response.Total = grandTotal
	response.Count = totalCount
	if totalCount > 0 {
		response.Average = grandTotal.Div(decimal.NewFromInt(totalCount))
	}

	return rows.Err()
}

func buildAnalyticsQuery(req *models.AnalyticsRequest) (string, []interface{}, error) {
	var builder strings.Builder
	args := []interface{}{req.From, req.To}
	argIndex := 3

	// Начинаем построение SELECT clause
	builder.WriteString("SELECT\n        ")

	// Добавляем поле группировки если нужно
	if req.GroupBy != "" {
		groupByField, err := getGroupByClause(req.GroupBy)
		if err != nil {
			return "", nil, err
		}
		builder.WriteString(groupByField)
		builder.WriteString(" as group,\n        ")
	}

	// Базовые агрегатные функции
	builder.WriteString(`COALESCE(SUM(amount), 0) as total,
        COALESCE(AVG(amount), 0) as average,
        COUNT(*) as count,
        PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY amount) as median,
        PERCENTILE_CONT(0.9) WITHIN GROUP (ORDER BY amount) as percentile_90`)

	// Добавляем MIN и MAX только для группировки
	if req.GroupBy != "" {
		builder.WriteString(",\n        MIN(amount) as min,\n        MAX(amount) as max")
	}

	// FROM и WHERE clauses
	builder.WriteString("\n        FROM sales\n        WHERE date BETWEEN $1 AND $2")

	// Добавляем фильтры
	if req.Category != "" {
		builder.WriteString(fmt.Sprintf(" AND category = $%d", argIndex))
		args = append(args, req.Category)
		argIndex++
	}

	if req.Type != "" {
		builder.WriteString(fmt.Sprintf(" AND type = $%d", argIndex))
		args = append(args, req.Type)
		argIndex++
	}

	// Добавляем GROUP BY и ORDER BY если есть группировка
	if req.GroupBy != "" {
		groupByField, _ := getGroupByClause(req.GroupBy) // Ошибка уже проверена выше
		builder.WriteString(fmt.Sprintf("\n        GROUP BY %s\n        ORDER BY %s", groupByField, groupByField))
	}

	return builder.String(), args, nil
}

func getGroupByClause(groupBy string) (string, error) {
	var groupByClause string
	switch groupBy {
	case "day":
		groupByClause = "DATE(date)"
	case "week":
		groupByClause = "EXTRACT(WEEK FROM date)"
	case "month":
		groupByClause = "EXTRACT(MONTH FROM date)"
	case "category":
		groupByClause = "category"
	default:
		return "", fmt.Errorf("invalid group_by value")
	}
	return groupByClause, nil
}
