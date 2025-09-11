package service

import (
	"context"

	"github.com/pozedorum/WB_project_3/task5/internal/models"
)

type Repository interface {
	CheckUserExistsByEmail(ctx context.Context, email string) (bool, error)
	CheckUserExistsByID(ctx context.Context, userID int) (bool, error)
	GetUserByID(ctx context.Context, userID int) (*models.UserInformation, error)

	CreateUser(ctx context.Context, user *models.UserInformation) error
	GetUserHash(ctx context.Context, email string) (string, error)

	CreateEvent(ctx context.Context, event *models.EventInformation) error
	GetEventByID(ctx context.Context, id int) (*models.EventInformation, error)

	CreateBookingWithSeatUpdate(ctx context.Context, booking *models.BookingInformation) error
	ConfirmBooking(ctx context.Context, bookingCode string) error
	GetExpiredBookings(ctx context.Context) ([]*models.BookingInformation, error)
	CancelBooking(ctx context.Context, bookingID int) error
}
