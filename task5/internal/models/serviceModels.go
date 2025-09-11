package models

import (
	"time"

	"github.com/wb-go/wbf/retry"
)

type UserInformation struct {
	ID           int       `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Name         string    `json:"name" db:"name"`
	Phone        string    `json:"phone" db:"phone"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type EventInformation struct {
	ID             int           `json:"id" db:"id"`
	Name           string        `json:"name" binding:"required" db:"name"`
	Date           time.Time     `json:"date" binding:"required" db:"total_seats"`
	Cost           int           `json:"cost" db:"cost"`
	TotalSeats     int           `json:"total_seats" binding:"required,min=1" db:"total_seats"`
	AvailableSeats int           `json:"-" db:"available_seats"`
	LifeSpan       time.Duration `json:"life_span" binding:"required" db:"booking_lifespan_minutes"`

	CreatedBy int       `json:"created_by" db:"created_by"` // ID пользователя-организатора
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type BookingInformation struct {
	ID          int        `json:"id" db:"id"`
	EventID     int        `json:"event_id" db:"event_id"`
	UserID      int        `json:"user_id" db:"user_id"`
	SeatCount   int        `json:"seat_count" db:"seat_count"`
	Status      string     `json:"status" db:"status"` // pending, confirmed, cancelled, expired
	BookingCode string     `json:"booking_code" db:"booking_code"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	ExpiresAt   time.Time  `json:"expires_at" db:"expires_at"`
	ConfirmedAt *time.Time `json:"confirmed_at,omitempty" db:"confirmed_at"`
}

var StandartStrategy = retry.Strategy{Attempts: 3, Delay: time.Second}
var ConsumerStrategy = retry.Strategy{Attempts: 5, Delay: 2 * time.Second}

const (
	StatusPending   = "pending"
	StatusConfirmed = "confirmed"
	StatusCancelled = "cancelled"
	StatusExpired   = "expired"

	StatusOK                  = 200
	StatusAccepted            = 202
	StatusBadRequest          = 400
	StatusNotFound            = 404
	StatusInternalServerError = 500
)
