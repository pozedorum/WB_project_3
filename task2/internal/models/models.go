package models

import (
	"time"

	"github.com/pozedorum/wbf/retry"
)

// URL модель
type ShortURL struct {
	ID          string    `json:"id"`
	ShortCode   string    `json:"short_code"`
	OriginalURL string    `json:"original_url"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at,omitempty"`
	IsActive    bool      `json:"is_active"`
	ClicksCount int       `json:"clicks"`
}

// Analytics модель
type ClickAnalyticsEntry struct {
	ID         string    `json:"id"`
	ShortURLID string    `json:"short_url_id"`
	UserAgent  string    `json:"user_agent"`
	IPAddress  string    `json:"ip_address"`
	Referrer   string    `json:"referrer"`
	CreatedAt  time.Time `json:"created_at"`
}

var StandartStrategy = retry.Strategy{Attempts: 3, Delay: time.Second}
var ConsumerStrategy = retry.Strategy{Attempts: 5, Delay: 2 * time.Second}

const (
	StatusOK                  = 200
	StatusAccepted            = 202
	StatusBadRequest          = 400
	StatusNotFound            = 404
	StatusInternalServerError = 500
)
