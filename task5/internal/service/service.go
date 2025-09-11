package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/pozedorum/WB_project_3/task5/internal/models"
	"github.com/wb-go/wbf/zlog"
	"golang.org/x/crypto/bcrypt"
)

type EventBookerService struct {
	repo Repository
}

func (servs *EventBookerService) RegisterUser(ctx context.Context, req *models.UserRegisterRequest) (*models.UserInformation, error) {

	exists, err := servs.repo.CheckUserExistsByEmail(ctx, req.Email)
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to check if user exists")
		return nil, err
	}
	if exists {
		zlog.Logger.Warn().Str("email", req.Email).Msg("user already exists")
		return nil, fmt.Errorf("user already exists")
	}
	pHash, err := HashPassword(req.Password)
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to hash password")
		return nil, err
	}
	newUser := &models.UserInformation{
		Email:        req.Email,
		PasswordHash: pHash,
		Name:         req.Name,
		Phone:        req.Phone,
	}
	err = servs.repo.CreateUser(ctx, newUser)
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to create user")
		return nil, err
	}
	return newUser, nil
}

func (servs *EventBookerService) AuthUser(ctx context.Context, req *models.UserLoginRequest) (*models.UserInformation, error) {

}

func (servs *EventBookerService) CreateEvent(ctx context.Context, req *models.EventRequest, userID int) (*models.EventInformation, error) {
	var err error
	if err = servs.CheckUserExistsByID(ctx, userID); err != nil {
		return nil, err
	}

	newEvent := &models.EventInformation{
		Name:           req.Name,
		Date:           req.Date,
		Cost:           req.Cost,
		TotalSeats:     req.TotalSeats,
		AvailableSeats: req.TotalSeats,
		LifeSpan:       req.LifeSpan,
		CreatedBy:      userID,
	}
	if err = servs.repo.CreateEvent(ctx, newEvent); err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to create event")
		return nil, err
	}

	return newEvent, nil
}

func (servs *EventBookerService) BookEvent(ctx context.Context, req *models.BookingRequest, userID int) (*models.BookingInformation, error) {
	var err error
	if err = servs.CheckUserExistsByID(ctx, userID); err != nil {
		return nil, err
	}

	if _, err = servs.repo.GetEventByID(ctx, req.EventID); err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to get event by id")
		return nil, err
	}
	booking := models.BookingInformation{
		EventID:     req.EventID,
		UserID:      userID,
		SeatCount:   req.SeatCount,
		Status:      models.StatusPending,
		BookingCode: uuid.NewString(),
	}
	if err = servs.repo.CreateBookingWithSeatUpdate(ctx, &booking); err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to create booking")
		return nil, err
	}

	return &booking, nil
}

func (servs *EventBookerService) CheckUserExistsByID(ctx context.Context, userID int) error {
	exists, err := servs.repo.CheckUserExistsByID(ctx, userID)
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("failed to check if user exists")
		return err
	}
	if !exists {
		zlog.Logger.Error().Int("user_id", userID).Msg("user does not exists")
		return fmt.Errorf("user does not exists")
	}
	return nil
}

func HashPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	// cost = 12 - хороший баланс между безопасностью и производительностью
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedBytes), nil
}

// CheckPassword проверяет соответствие пароля и хеша
func CheckPassword(password, hashedPassword string) bool {
	if password == "" || hashedPassword == "" {
		return false
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
