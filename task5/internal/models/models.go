package models

import (
	"time"

	"github.com/wb-go/wbf/retry"
)

type EventInformation struct {
	Name       string        `json:"name" binding:"required"`
	Date       time.Time     `json:"date" binding:"required"`
	Cost       int           `json:"cost"`
	NumOfSeats int           `json:"num_of_seats" binding:"required,min=1"`
	LifeSpan   time.Duration `json:"life_span" binding:"required"`
}

type BookingRequest struct {
	Name string `json:"name" binding:"required"`
	Seat int    `json:"seat" binding:"required"`
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
