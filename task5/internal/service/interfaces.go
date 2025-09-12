package service

import (
	"context"

	"github.com/pozedorum/WB_project_3/task5/internal/models"
)

type Repository interface {
	GetUserByID(ctx context.Context, userID int) (*models.UserInformation, error)
	GetUserByEmail(ctx context.Context, email string) (*models.UserInformation, error)
	GetUserHash(ctx context.Context, email string) (string, error)
	GetEventByID(ctx context.Context, id int) (*models.EventInformation, error)
	GetAllEvents(ctx context.Context) ([]*models.EventInformation, error)
	GetBookingByCode(ctx context.Context, bookingCode string) (*models.BookingInformation, error)

	CreateUser(ctx context.Context, user *models.UserInformation) error
	CreateEvent(ctx context.Context, event *models.EventInformation) error
	CreateBookingWithSeatUpdate(ctx context.Context, booking *models.BookingInformation) error

	CheckEmailExists(ctx context.Context, email string) (bool, error)
	ConfirmBooking(ctx context.Context, bookingCode string) error
	GetExpiredBookings(ctx context.Context) ([]*models.BookingInformation, error)
	CancelBooking(ctx context.Context, bookingID int) error
}
