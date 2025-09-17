package models

import (
	"time"

	"github.com/shopspring/decimal"
	"github.com/wb-go/wbf/retry"
)

// SaleInformation -- cтруктура продажи с полной информацией
type SaleInformation struct {
	ID          int64           `json:"id" db:"id"`
	Amount      decimal.Decimal `json:"amount" db:"amount" validate:"required,gt=0"`
	Type        string          `json:"type" db:"type" validate:"required,oneof=income expense"`
	Category    string          `json:"category" db:"category" validate:"required"`
	Description string          `json:"description" db:"description"`
	Date        time.Time       `json:"date" db:"date" validate:"required"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
}

// SaleRequest -- структура запроса и создании или изменении записи в таблице
type SaleRequest struct {
	Amount      decimal.Decimal
	Type        string
	Category    string
	Description string
	Date        time.Time
}

// AnalyticsRequest -- структура запроса аналитики
type AnalyticsRequest struct {
	From     time.Time `form:"from" json:"from" validate:"required"`
	To       time.Time `form:"to" json:"to" validate:"required"`
	Type     string    `form:"type" json:"type"`                                                  // income/expense
	Category string    `form:"category" json:"category"`                                          // фильтр по категории
	GroupBy  string    `form:"group_by" json:"group_by" validate:"oneof=day week month category"` // группировка
}

type AnalyticsResponse struct {
	Total        decimal.Decimal   `json:"total" db:"total"`
	Average      decimal.Decimal   `json:"average" db:"average"`
	Count        int64             `json:"count" db:"count"`
	Median       decimal.Decimal   `json:"median" db:"median"`
	Percentile90 decimal.Decimal   `json:"percentile_90" db:"percentile_90"`
	GroupedData  []GroupedDataItem `json:"grouped_data,omitempty"`
}

type SalesSummaryResponse struct {
	SumAmount     decimal.Decimal `json:"sum_amount" db:"sum_amount"`
	ItemsCount    int64           `json:"items_count" db:"items_count"`
	AverageAmount decimal.Decimal `json:"average_count" db:"average_count"`
}

type GroupedDataItem struct {
	Group        string          `json:"group" db:"group"`
	Total        decimal.Decimal `json:"total" db:"total"`
	Average      decimal.Decimal `json:"average" db:"average"`
	Count        int64           `json:"count" db:"count"`
	Median       decimal.Decimal `json:"median,omitempty" db:"median"`
	Percentile90 decimal.Decimal `json:"percentile_90,omitempty" db:"percentile_90"`
	Min          decimal.Decimal `json:"min,omitempty" db:"min"`
	Max          decimal.Decimal `json:"max,omitempty" db:"max"`
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
