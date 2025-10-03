package models

import (
	"time"

	"github.com/wb-go/wbf/retry"
)

// Item - товар на складе
type Item struct {
	ID        int64     `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Price     int64     `json:"price" db:"price"` // цена в копейках/центах
	CreatedBy string    `json:"created_by" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ItemRequest - запрос на создание/изменение товара
type ItemRequest struct {
	Name  string `json:"name" binding:"required"`
	Price int64  `json:"price" binding:"required,gt=0"`
}

// HistoryRecord - запись в истории изменений
type HistoryRecord struct {
	ID        int64     `json:"id" db:"id"`
	ItemID    int64     `json:"item_id" db:"item_id"`
	Action    string    `json:"action" db:"action"` // CREATE, UPDATE, DELETE
	ChangedBy string    `json:"changed_by" db:"changed_by"`
	ChangedAt time.Time `json:"changed_at" db:"changed_at"`
}

// JWTClaims - данные в токене
type JWTClaims struct {
	Username string   `json:"username"`
	Role     UserRole `json:"role"`
	Exp      int64    `json:"exp"`
}

// UserRole - роли пользователей
type UserRole string

const (
	RoleAdmin   UserRole = "admin"
	RoleManager UserRole = "manager"
	RoleViewer  UserRole = "viewer"
)

var (
	StandardStrategy = retry.Strategy{Attempts: 3, Delay: time.Second}
	ConsumerStrategy = retry.Strategy{Attempts: 5, Delay: 2 * time.Second}
)

const (
	StatusOK                  = 200
	StatusAccepted            = 202
	StatusBadRequest          = 400
	StatusNotFound            = 404
	StatusInternalServerError = 500
)
