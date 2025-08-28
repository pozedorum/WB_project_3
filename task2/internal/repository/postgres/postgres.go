package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/pozedorum/WB_project_3/task2/internal/models"
	"github.com/pozedorum/WB_project_3/task2/internal/utils"

	"github.com/pozedorum/wbf/dbpg"
	"github.com/pozedorum/wbf/zlog"
)

type ShortURLRepository struct {
	db *dbpg.DB
}

func NewShortURLRepositoryWithDB(masterDSN string, slaveDSNs []string, opts *dbpg.Options) (*ShortURLRepository, error) {
	db, err := dbpg.New(masterDSN, slaveDSNs, opts)
	if err != nil {
		return nil, err
	}
	return NewShortURLRepository(db), nil
}

func NewShortURLRepository(db *dbpg.DB) *ShortURLRepository {
	return &ShortURLRepository{db: db}
}

// Создание и обновление
func (sr *ShortURLRepository) CreateShortURL(ctx context.Context, n *models.ShortURL) error {
	createQuery := `INSERT INTO short_urls (short_code, original_url, created_at, clicks_count) 
		VALUES ($1, $2, $3, $4)`

	_, err := sr.db.ExecWithRetry(ctx, models.StandardStrategy, createQuery,
		n.ShortCode, n.OriginalURL, n.CreatedAt, n.ClicksCount)

	if err != nil {
		zlog.Logger.Error().Err(err).Str("short_code", n.ShortCode).Msg("Failed to create url in database")
	} else {
		zlog.Logger.Info().Str("short_code", n.ShortCode).Msg("URL created in database")
	}

	return err
}

func (sr *ShortURLRepository) RegisterClick(ctx context.Context, click *models.ClickAnalyticsEntry) error {
	// Парсим User-Agent для детальной аналитики
	userAgentInfo := utils.ParseUserAgent(click.UserAgent)

	// Начинаем транзакцию с контекстом
	tx, err := sr.db.Master.BeginTx(ctx, nil)
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("Failed to begin transaction")
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Запрос для вставки данных о клике
	// ВАЖНО: Используем только базовые поля, которые есть в таблице из миграции 001
	query := `INSERT INTO url_clicks 
        (short_url_id, user_agent,ip_address, created_at) 
        SELECT id, $2, $3, $4
        FROM short_urls WHERE short_code = $1
        RETURNING short_url_id`

	var shortURLID int
	err = tx.QueryRowContext(ctx, query,
		click.ShortCode,
		click.UserAgent,
		click.IPAddress,
		click.CreatedAt,
	).Scan(&shortURLID)
	// Обрабатываем ошибку
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			zlog.Logger.Warn().Str("short_code", click.ShortCode).Msg("Attempted to register click for non-existent short code")
			return models.ErrShortURLNotFound
		}
		zlog.Logger.Error().Err(err).Str("short_code", click.ShortCode).Msg("Failed to insert click analytics")
		return fmt.Errorf("database error on click insert: %w", err)
	}

	// Обновляем счетчик кликов
	_, err = tx.ExecContext(ctx,
		"UPDATE short_urls SET clicks_count = clicks_count + 1 WHERE id = $1",
		shortURLID,
	)
	if err != nil {
		zlog.Logger.Error().Err(err).Int("url_id", shortURLID).Msg("Failed to update click count")
		return fmt.Errorf("database error on count update: %w", err)
	}

	// Коммитим транзакцию
	if err := tx.Commit(); err != nil {
		zlog.Logger.Error().Err(err).Msg("Failed to commit transaction for click registration")
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Логируем аналитику (данные есть в userAgentInfo, но не сохраняем в БД)
	zlog.Logger.Debug().
		Int("url_id", shortURLID).
		Str("short_code", click.ShortCode).
		Str("browser", userAgentInfo.Browser).
		Str("os", userAgentInfo.OS).
		Str("device", userAgentInfo.Device).
		Msg("Click registered successfully")

	return nil
}

// Чтение

func (sr *ShortURLRepository) GetOriginalURLIfExists(ctx context.Context, shortCode string) (*models.ShortURL, error) {
	var shortURL models.ShortURL

	err := sr.db.Master.QueryRowContext(ctx,
		`SELECT id, short_code, original_url, created_at, clicks_count 
		 FROM short_urls WHERE short_code = $1`,
		shortCode,
	).Scan(
		&shortURL.ID,
		&shortURL.ShortCode,
		&shortURL.OriginalURL,
		&shortURL.CreatedAt,
		&shortURL.ClicksCount,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrShortURLNotFound
		}
		zlog.Logger.Error().Err(err).Str("short_code", shortCode).Msg("Failed to get short URL")
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &shortURL, nil
}

// internal/repository/postgres.go
func (sr *ShortURLRepository) GetStatisticsByShortCode(ctx context.Context, shortCode string, period string, groupBy string) (*models.AnalyticsResponse, error) {
	var shortURLID int
	err := sr.db.Master.QueryRowContext(ctx,
		"SELECT id FROM short_urls WHERE short_code = $1",
		shortCode,
	).Scan(&shortURLID)
	if err != nil {
		return nil, err
	}

	analytics := &models.AnalyticsResponse{}

	// baseCondition нужен для правильных фильтров
	baseCondition := "WHERE short_url_id = $1"
	periodFilter := getPeriodFilter(period)
	if periodFilter != "" {
		baseCondition += " AND " + periodFilter
	}

	err = sr.db.Master.QueryRowContext(ctx, `
        SELECT 
            COUNT(*) as total_clicks,
            COUNT(DISTINCT ip_address) as unique_visitors
        FROM url_clicks `+baseCondition, shortURLID).Scan(&analytics.TotalClicks, &analytics.UniqueVisitors)
	if err != nil {
		return nil, err
	}

	// Обрабатываем группировку
	switch groupBy {
	case "day":
		analytics.DailyStats = sr.getDailyStats(ctx, shortURLID, period)
	case "month":
		analytics.MonthlyStats = sr.getMonthlyStats(ctx, shortURLID, period)
	case "user-agent":
		analytics.UserAgentStats = sr.getUserAgentStats(ctx, shortURLID, period)
	case "browser":
		analytics.BrowserStats = sr.getBrowserStats(ctx, shortURLID, period)
	case "os":
		analytics.OSStats = sr.getOSStats(ctx, shortURLID, period)
	case "device":
		analytics.DeviceStats = sr.getDeviceStats(ctx, shortURLID, period)
	default:
		// По умолчанию возвращаем все виды статистики
		analytics.DailyStats = sr.getDailyStats(ctx, shortURLID, period)
		analytics.MonthlyStats = sr.getMonthlyStats(ctx, shortURLID, period)
		analytics.UserAgentStats = sr.getUserAgentStats(ctx, shortURLID, period)
		analytics.BrowserStats = sr.getBrowserStats(ctx, shortURLID, period)
		analytics.OSStats = sr.getOSStats(ctx, shortURLID, period)
		analytics.DeviceStats = sr.getDeviceStats(ctx, shortURLID, period)
	}

	return analytics, nil
}

// Вспомогательная функция для фильтра по периоду

func getPeriodFilter(period string) string {
	switch period {
	case "1d":
		return "created_at >= NOW() - INTERVAL '1 day'"
	case "7d":
		return "created_at >= NOW() - INTERVAL '7 days'"
	case "30d":
		return "created_at >= NOW() - INTERVAL '30 days'"
	case "today":
		return "created_at >= CURRENT_DATE"
	case "yesterday":
		return "created_at >= CURRENT_DATE - INTERVAL '1 day' AND created_at < CURRENT_DATE"
	default:
		return "" // Все время
	}
}

// Обновляем методы статистики для поддержки периода
func (sr *ShortURLRepository) getDailyStats(ctx context.Context, shortURLID int, period string) []models.DailyStat {
	baseCondition := "WHERE short_url_id = $1"
	periodFilter := getPeriodFilter(period)
	if periodFilter != "" {
		baseCondition += " AND " + periodFilter
	}

	query := `
        SELECT 
            DATE(created_at) as date,
            COUNT(*) as clicks,
            COUNT(DISTINCT ip_address) as unique_ips
        FROM url_clicks 
        ` + baseCondition + `
        GROUP BY DATE(created_at)
        ORDER BY date DESC
        LIMIT 30
    `

	rows, err := sr.db.Master.QueryContext(ctx, query, shortURLID)
	if err != nil {
		zlog.Logger.Error().Err(err).Int("url_id", shortURLID).Msg("Failed to get daily stats")
		return nil
	}
	defer rows.Close()

	var stats []models.DailyStat
	for rows.Next() {
		var stat models.DailyStat
		if err := rows.Scan(&stat.Date, &stat.Clicks, &stat.UniqueIPs); err != nil {
			continue
		}
		stats = append(stats, stat)
	}
	return stats
}

func (sr *ShortURLRepository) getMonthlyStats(ctx context.Context, shortURLID int, period string) []models.MonthlyStat {
	baseCondition := "WHERE short_url_id = $1"
	periodFilter := getPeriodFilter(period)
	if periodFilter != "" {
		baseCondition += " AND " + periodFilter
	}

	query := `
        SELECT 
            TO_CHAR(created_at, 'YYYY-MM') as month,
            COUNT(*) as clicks,
            COUNT(DISTINCT ip_address) as unique_ips
        FROM url_clicks 
        ` + baseCondition + `
        GROUP BY TO_CHAR(created_at, 'YYYY-MM')
        ORDER BY month DESC
    `

	rows, err := sr.db.Master.QueryContext(ctx, query, shortURLID)
	if err != nil {
		zlog.Logger.Error().Err(err).Int("url_id", shortURLID).Msg("Failed to get monthly stats")
		return []models.MonthlyStat{}
	}
	defer rows.Close()

	var stats []models.MonthlyStat
	for rows.Next() {
		var stat models.MonthlyStat
		if err := rows.Scan(&stat.Month, &stat.Clicks, &stat.UniqueIPs); err != nil {
			continue
		}
		stats = append(stats, stat)
	}
	return stats
}

func (sr *ShortURLRepository) getUserAgentStats(ctx context.Context, shortURLID int, period string) []models.UserAgentStat {
	baseCondition := "WHERE short_url_id = $1 AND user_agent IS NOT NULL"
	periodFilter := getPeriodFilter(period)
	if periodFilter != "" {
		baseCondition += " AND " + periodFilter
	}

	query := `
        SELECT 
            user_agent,
            COUNT(*) as count
        FROM url_clicks 
        ` + baseCondition + `
        GROUP BY user_agent
        ORDER BY count DESC
        LIMIT 20
    `

	rows, err := sr.db.Master.QueryContext(ctx, query, shortURLID)
	if err != nil {
		zlog.Logger.Error().Err(err).Int("url_id", shortURLID).Msg("Failed to get user agent stats")
		return nil
	}
	defer rows.Close()

	var stats []models.UserAgentStat
	for rows.Next() {
		var stat models.UserAgentStat
		if err := rows.Scan(&stat.UserAgent, &stat.Count); err != nil {
			continue
		}
		stats = append(stats, stat)
	}
	return stats
}

func (sr *ShortURLRepository) getBrowserStats(ctx context.Context, shortURLID int, period string) []models.BrowserStat {
	baseCondition := "WHERE short_url_id = $1 AND user_agent IS NOT NULL"
	periodFilter := getPeriodFilter(period)
	if periodFilter != "" {
		baseCondition += " AND " + periodFilter
	}

	query := `
        SELECT user_agent FROM url_clicks 
        ` + baseCondition + `
    `

	rows, err := sr.db.Master.QueryContext(ctx, query, shortURLID)
	if err != nil {
		zlog.Logger.Error().Err(err).Int("url_id", shortURLID).Msg("Failed to get browser stats")
		return nil
	}
	defer rows.Close()

	browserCounts := make(map[string]int64)
	for rows.Next() {
		var userAgent string
		if err := rows.Scan(&userAgent); err != nil {
			continue
		}
		browser := utils.ExtractBrowser(userAgent)
		browserCounts[browser]++
	}

	var stats []models.BrowserStat
	for browser, count := range browserCounts {
		stats = append(stats, models.BrowserStat{
			Browser: browser,
			Count:   count,
		})
	}
	return stats
}

func (sr *ShortURLRepository) getOSStats(ctx context.Context, shortURLID int, period string) []models.OSStat {
	baseCondition := "WHERE short_url_id = $1 AND user_agent IS NOT NULL"
	periodFilter := getPeriodFilter(period)
	if periodFilter != "" {
		baseCondition += " AND " + periodFilter
	}

	query := `
        SELECT user_agent FROM url_clicks 
        ` + baseCondition + `
    `

	rows, err := sr.db.Master.QueryContext(ctx, query, shortURLID)
	if err != nil {
		zlog.Logger.Error().Err(err).Int("url_id", shortURLID).Msg("Failed to get OS stats")
		return nil
	}
	defer rows.Close()

	osCounts := make(map[string]int64)
	for rows.Next() {
		var userAgent string
		if err := rows.Scan(&userAgent); err != nil {
			continue
		}
		os := utils.ExtractOS(userAgent)
		osCounts[os]++
	}

	var stats []models.OSStat
	for os, count := range osCounts {
		stats = append(stats, models.OSStat{
			OS:    os,
			Count: count,
		})
	}
	return stats
}

func (sr *ShortURLRepository) getDeviceStats(ctx context.Context, shortURLID int, period string) []models.DeviceStat {
	baseCondition := "WHERE short_url_id = $1 AND user_agent IS NOT NULL"
	periodFilter := getPeriodFilter(period)
	if periodFilter != "" {
		baseCondition += " AND " + periodFilter
	}

	query := `
        SELECT user_agent FROM url_clicks 
        ` + baseCondition + `
    `

	rows, err := sr.db.Master.QueryContext(ctx, query, shortURLID)
	if err != nil {
		zlog.Logger.Error().Err(err).Int("url_id", shortURLID).Msg("Failed to get device stats")
		return nil
	}
	defer rows.Close()

	deviceCounts := make(map[string]int64)
	for rows.Next() {
		var userAgent string
		if err := rows.Scan(&userAgent); err != nil {
			continue
		}
		device := utils.ExtractDevice(userAgent)
		deviceCounts[device]++
	}

	var stats []models.DeviceStat
	for device, count := range deviceCounts {
		stats = append(stats, models.DeviceStat{
			Device: device,
			Count:  count,
		})
	}
	return stats
}

func (sr *ShortURLRepository) Close() {
	sr.db.Master.Close()
	for _, slave := range sr.db.Slaves {
		if slave != nil {
			slave.Close()
		}
	}
	zlog.Logger.Info().Msg("PostgreSQL connections closed")
}

// Implement Repository interface
// var _ service.Repository = (*ShortURLRepository)(nil)
