package models

import (
	"time"

	"github.com/pozedorum/wbf/retry"
)

type Notification struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Message   string    `json:"message"`
	Channel   string    `json:"channel"` // email, telegram
	SendAt    time.Time `json:"send_at"`
	Status    string    `json:"status"` // pending, sent, failed, canceled
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateNotificationRequest struct {
	UserID  string    `json:"user_id" binding:"required"`
	Message string    `json:"message" binding:"required"`
	Channel string    `json:"channel" binding:"required"`
	SendAt  time.Time `json:"send_at" binding:"required"`
}

type NotificationResponse struct {
	ID      string    `json:"id"`
	Status  string    `json:"status"`
	Channel string    `json:"channel"` // email, telegram
	Message string    `json:"message"`
	SendAt  time.Time `json:"send_at"`
}

var StandartStrategy = retry.Strategy{Attempts: 3, Delay: time.Second}
var ConsumerStrategy = retry.Strategy{Attempts: 5, Delay: 2 * time.Second}

const (
	StatusPending  = "pending"
	StatusSent     = "sent"
	StatusFailed   = "failed"
	StatusCanceled = "canceled"

	StatusOK                  = 200
	StatusAccepted            = 202
	StatusBadRequest          = 400
	StatusNotFound            = 404
	StatusInternalServerError = 500
)
