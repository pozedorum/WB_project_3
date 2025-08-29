package models

import (
	"errors"
	"time"

	"github.com/pozedorum/wbf/retry"
)

// URL модель
type ShortURL struct {
	ID          int       `json:"id" db:"id"`                 // SERIAL PRIMARY KEY
	ShortCode   string    `json:"short_code" db:"short_code"` // VARCHAR(10) UNIQUE
	OriginalURL string    `json:"original_url" db:"original_url"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	ClicksCount int       `json:"clicks_count" db:"clicks_count"`
}

// Модель информации о клике
type ClickAnalyticsEntry struct {
	ID        string    `json:"id"`         // Опционально, может генерироваться БД
	ShortCode string    `json:"short_code"` // Публичный код ссылки (abc123)
	UserAgent string    `json:"user_agent"`
	IPAddress string    `json:"ip_address"`
	CreatedAt time.Time `json:"created_at"`
}

// Модель собранной аналитики
type AnalyticsResponse struct {
	TotalClicks    int64            `json:"total_clicks"`
	UniqueVisitors int64            `json:"unique_visitors"`
	DailyStats     []DailyStat      `json:"daily_stats,omitempty"`
	MonthlyStats   []MonthlyStat    `json:"monthly_stats,omitempty"`
	UserAgentStats []UserAgentStat  `json:"user_agent_stats,omitempty"`
	BrowserStats   []BrowserStat    `json:"browser_stats,omitempty"`
	OSStats        []OSStat         `json:"os_stats,omitempty"`
	DeviceStats    []DeviceStat     `json:"device_stats,omitempty"`
	TimeSeries     []TimeSeriesStat `json:"time_series,omitempty"`
}

type UserAgentInfo struct {
	Browser string `json:"browser"`
	OS      string `json:"os"`
	Device  string `json:"device"`
}

type DailyStat struct {
	Date      string `json:"date"`
	Clicks    int64  `json:"clicks"`
	UniqueIPs int64  `json:"unique_ips"`
}

type MonthlyStat struct {
	Month     string `json:"month"`
	Clicks    int64  `json:"clicks"`
	UniqueIPs int64  `json:"unique_ips"`
}

type UserAgentStat struct {
	UserAgent string `json:"user_agent"`
	Count     int64  `json:"count"`
}

type BrowserStat struct {
	Browser string `json:"browser"`
	Count   int64  `json:"count"`
}

type OSStat struct {
	OS    string `json:"os"`
	Count int64  `json:"count"`
}

type DeviceStat struct {
	Device string `json:"device"`
	Count  int64  `json:"count"`
}

type TimeSeriesStat struct {
	Timestamp time.Time `json:"timestamp"`
	Clicks    int64     `json:"clicks"`
}

var (
	StandardStrategy = retry.Strategy{Attempts: 3, Delay: time.Second}
	ConsumerStrategy = retry.Strategy{Attempts: 5, Delay: 2 * time.Second}
)

var (
	ErrShortURLNotFound   = errors.New("short URL not found")
	ErrDuplicateShortCode = errors.New("duplicate short code")
)

const (
	StatusOK                  = 200
	StatusAccepted            = 202
	StatusFound               = 302
	StatusBadRequest          = 400
	StatusNotFound            = 404
	StatisConflict            = 409
	StatusInternalServerError = 500
)
