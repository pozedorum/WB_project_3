package models

import (
	"time"
)

type UserRegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
	Phone    string `json:"phone"`
}

type UserLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type EventRequest struct {
	Name       string    `json:"name" binding:"required"`
	Date       time.Time `json:"date" binding:"required"`
	Cost       int       `json:"cost"`
	TotalSeats int       `json:"total_seats" binding:"required,min=1"`
	LifeSpan   string    `json:"life_span" binding:"required"`
}

type BookingRequest struct {
	EventID   int       `json:"event_id"` // будет извлекаться из самого запроса http://localhost:8080/events/1/book
	SeatCount int       `json:"seat_count" binding:"required,min=1"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	// UserID будет извлекаться из контекста (JWT токена)

}

type ConfirmBookingRequest struct {
	BookingCode string `json:"booking_code" binding:"required"`
}
