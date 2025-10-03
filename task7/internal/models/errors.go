package models

import (
	"database/sql"
	"errors"
)

var (
	ErrItemEmptyName    = errors.New("empty item name")
	ErrItemInvalidPrice = errors.New("item price is negative or null")
	ErrInvalidID        = errors.New("id is negative or null")

	ErrNoRows = sql.ErrNoRows

	ErrNotEnoughRights = errors.New("not enough rights")
)
