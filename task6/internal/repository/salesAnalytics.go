package repository

import (
	"bytes"
	"context"
	"database/sql" // нужен только для sql.Row
	"encoding/csv"
	"fmt"
	"strconv"
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
	}, models.StandardStrategy)
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
	}, models.StandardStrategy)
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
	}, models.StandardStrategy)
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
	// Если группировка не указана - экспортируем сырые данные
	if req.GroupBy == "" {
		// Получаем сырые данные
		query := `
            SELECT id, amount, type, category, description, date, created_at, updated_at
            FROM sales WHERE date BETWEEN $1 AND $2
        `
		args := []interface{}{req.From, req.To}
		argIndex := 3

		if req.Category != "" {
			query += fmt.Sprintf(" AND category = $%d", argIndex)
			args = append(args, req.Category)
			argIndex++
		}

		if req.Type != "" {
			query += fmt.Sprintf(" AND type = $%d", argIndex)
			args = append(args, req.Type)
			argIndex++
		}

		query += " ORDER BY date DESC"

		var sales []models.SaleInformation
		rows, err := repo.db.Master.QueryContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var sale models.SaleInformation
			err := rows.Scan(
				&sale.ID, &sale.Amount, &sale.Type, &sale.Category,
				&sale.Description, &sale.Date, &sale.CreatedAt, &sale.UpdatedAt,
			)
			if err != nil {
				return nil, err
			}
			sales = append(sales, sale)
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return salesToCSV(sales)
	}

	// Если есть группировка - используем аналитику и преобразуем в CSV
	analytics, err := repo.GetAnalytics(ctx, req)
	if err != nil {
		return nil, err
	}

	return analyticsToCSV(analytics, req.GroupBy)
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
	}, models.StandardStrategy)

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

func analyticsToCSV(analytics *models.AnalyticsResponse, groupBy string) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	if err := writer.Write([]string{"STATISTICS SUMMARY"}); err != nil {
		return nil, err
	}

	if err := writer.Write([]string{
		"Total: " + analytics.Total.String(),
		"Average: " + analytics.Average.String(),
		"Count: " + strconv.FormatInt(analytics.Count, 10),
		"Median: " + analytics.Median.String(),
		"90th Percentile: " + analytics.Percentile90.String(),
	}); err != nil {
		return nil, err
	}

	if err := writer.Write([]string{}); err != nil {
		return nil, err
	}

	headers := make([]string, 0, 8)
	if groupBy == "category" {
		headers = append(headers, "Category")
	} else {
		headers = append(headers, "Period")
	}
	headers = append(headers, []string{"Total", "Average", "Count", "Median", "90th Percentile", "Min", "Max"}...)

	if err := writer.Write(headers); err != nil {
		return nil, err
	}

	if len(analytics.GroupedData) == 0 {
		if err := writer.Write([]string{"No data available for the selected period"}); err != nil {
			return nil, err
		}
	}
	for _, group := range analytics.GroupedData {

		groupName := formatGroupName(group.Group, groupBy)

		record := []string{
			groupName,
			group.Total.String(),
			group.Average.String(),
			strconv.FormatInt(group.Count, 10),
			group.Median.String(),
			group.Percentile90.String(),
			group.Min.String(),
			group.Max.String(),
		}

		if err := writer.Write(record); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func salesToCSV(sales []models.SaleInformation) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Заголовок CSV
	headers := []string{"ID", "Amount", "Type", "Category", "Description", "Date", "CreatedAt", "UpdatedAt"}
	if err := writer.Write(headers); err != nil {
		return nil, err
	}

	// Данные
	for _, sale := range sales {
		record := []string{
			strconv.FormatInt(sale.ID, 10),
			sale.Amount.String(),
			sale.Type,
			sale.Category,
			sale.Description,
			sale.Date.Format("2006-01-02 15:04:05"),
			sale.CreatedAt.Format("2006-01-02 15:04:05"),
			sale.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		if err := writer.Write(record); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func formatGroupName(groupValue string, groupBy string) string {
	switch groupBy {
	case "day":
		if date, err := time.Parse("2006-01-02", groupValue); err == nil {
			return date.Format("January 02, 2006")
		}
	case "week":
		return "Week " + groupValue
	case "month":
		if monthNum, err := strconv.Atoi(groupValue); err == nil && monthNum >= 1 && monthNum <= 12 {
			return time.Month(monthNum).String()
		}
	case "category":
		return groupValue
	}
	return groupValue
}
