package models

import "errors"

var (
	ErrUserAlreadyRegistered   = errors.New("user already registered")
	ErrUserNotFound            = errors.New("user not found")
	ErrWrongPassword           = errors.New("wrong password")
	ErrEventNotFound           = errors.New("event not found")
	ErrBookingNotFound         = errors.New("booking not found")
	ErrNotEnoughAvailableSeats = errors.New("not enough seats available")
)
