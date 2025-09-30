package models

import "errors"

var (
	ErrInvalidItemID   = errors.New("invalid item ID")
	ErrNegativeAmount  = errors.New("amount must be positive")
	ErrInvalidType     = errors.New("type must be 'income' or 'expense'")
	ErrInvalidGroupBy  = errors.New("group_by must be 'day', 'week', 'month', or 'category'")
	ErrEmptyCategory   = errors.New("category is required")
	ErrEmptyDate       = errors.New("date is required")
	ErrEmptyFromToDate = errors.New("both 'from' and 'to' parameters are required")
	ErrWrongTimeRange  = errors.New("'from' date cannot be after 'to' date")
	ErrFutureDate      = errors.New("date cannot be in the future")

	ErrSaleNotFound  = errors.New("sale not found")
	ErrNegativeIndex = errors.New("invalid ID: must be positive integer")
)
