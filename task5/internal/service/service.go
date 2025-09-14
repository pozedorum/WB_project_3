package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/pozedorum/WB_project_3/task5/internal/models"
	"github.com/pozedorum/WB_project_3/task5/pkg/logger"
	"github.com/wb-go/wbf/zlog"
	"golang.org/x/crypto/bcrypt"
)

type EventBookerService struct {
	repo Repository
}

func NewEventBookerService(repo Repository) *EventBookerService {
	return &EventBookerService{repo: repo}
}

func (servs *EventBookerService) RegisterUser(ctx context.Context, req *models.UserRegisterRequest) (*models.UserInformation, error) {

	var err error
	if _, err = servs.repo.GetUserByEmail(ctx, req.Email); err == nil {
		return nil, models.ErrUserAlreadyRegistered
	}
	if err != nil && err != models.ErrUserNotFound {
		return nil, err
	}
	pHash, err := hashPassword(req.Password)
	if err != nil {
		logger.LogService(func() { zlog.Logger.Error().Err(err).Msg("failed to hash password") })
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
		logger.LogService(func() { zlog.Logger.Error().Err(err).Msg("failed to create user") })
		return nil, err
	}
	return newUser, nil
}

func (servs *EventBookerService) AuthUser(ctx context.Context, req *models.UserLoginRequest) (*models.UserInformation, error) {
	var dbUser *models.UserInformation
	var err error
	if dbUser, err = servs.repo.GetUserByEmail(ctx, req.Email); err != nil {
		return nil, err
	}

	if !checkPassword(req.Password, dbUser.PasswordHash) {
		logger.LogService(func() { zlog.Logger.Warn().Msg("password hashes do not match") })
		return nil, models.ErrWrongPassword
	}

	return dbUser, nil
}

func (servs *EventBookerService) CreateEvent(ctx context.Context, req *models.EventRequest, userID int) (*models.EventInformation, error) {
	var err error
	if _, err = servs.CheckUserExistsByID(ctx, userID); err != nil {
		return nil, err
	}
	lifespan, err := time.ParseDuration(req.LifeSpan)
	if err != nil {
		return nil, fmt.Errorf("invalid life_span format: %w", err)
	}
	logger.LogService(func() {
		zlog.Logger.Info().Str("str_event_lifespan", req.LifeSpan).Dur("event_lifespan", lifespan).Msg("event_lifespan!!!!!!")
	})
	newEvent := &models.EventInformation{
		Name:           req.Name,
		Date:           req.Date,
		Cost:           req.Cost,
		TotalSeats:     req.TotalSeats,
		AvailableSeats: req.TotalSeats,
		LifeSpan:       lifespan,
		CreatedBy:      userID,
	}
	if err = servs.repo.CreateEvent(ctx, newEvent); err != nil {
		logger.LogService(func() { zlog.Logger.Error().Err(err).Msg("failed to create event") })
		return nil, err
	}

	return newEvent, nil
}

func (servs *EventBookerService) GetEventInformation(ctx context.Context, eventID string) (*models.EventInformation, error) {
	// Конвертируем строковый ID в числовой
	id, err := strconv.Atoi(eventID)
	if err != nil {
		logger.LogService(func() {
			zlog.Logger.Error().
				Err(err).
				Str("event_id", eventID).
				Msg("Failed to convert event ID to integer")
		})
		return nil, fmt.Errorf("invalid event ID format")
	}

	// Получаем информацию о мероприятии из репозитория
	event, err := servs.repo.GetEventByID(ctx, id)
	if err != nil {
		if err == models.ErrEventNotFound {
			logger.LogService(func() {
				zlog.Logger.Warn().
					Err(err).
					Int("event_id", id).
					Msg("Event not found")
			})
			return nil, models.ErrEventNotFound
		}

		logger.LogService(func() {
			zlog.Logger.Error().
				Err(err).
				Int("event_id", id).
				Msg("Failed to get event information from repository")
		})
		return nil, fmt.Errorf("failed to get event information: %w", err)
	}

	logger.LogService(func() {
		zlog.Logger.Info().
			Int("event_id", id).
			Str("event_name", event.Name).
			Int("available_seats", event.AvailableSeats).
			Msg("Successfully retrieved event information")
	})

	return event, nil
}

func (servs *EventBookerService) GetAllEvents(ctx context.Context) ([]*models.EventInformation, error) {
	events, err := servs.repo.GetAllEvents(ctx)
	if err != nil {
		logger.LogService(func() {
			zlog.Logger.Error().Err(err).Msg("Failed to get all events")
		})
		return nil, err
	}

	logger.LogService(func() {
		zlog.Logger.Info().
			Int("events_count", len(events)).
			Msg("Successfully retrieved all events")
	})

	return events, nil
}

func (servs *EventBookerService) BookEvent(ctx context.Context, req *models.BookingRequest, userID int) (*models.BookingInformation, error) {
	var err error
	var event *models.EventInformation
	if _, err = servs.CheckUserExistsByID(ctx, userID); err != nil {
		return nil, err
	}

	if event, err = servs.repo.GetEventByID(ctx, req.EventID); err != nil {
		logger.LogService(func() { zlog.Logger.Error().Err(err).Msg("failed to get event by id") })
		return nil, err
	}

	logger.LogService(func() {
		zlog.Logger.Info().
			Str("event_name", event.Name).
			Dur("event_lifespan", event.LifeSpan).
			Int("lifespan_minutes", int(event.LifeSpan.Minutes())).
			Msg("Retrieved event for booking")
	})
	startTime := time.Now()
	expiresAt := startTime.Add(event.LifeSpan)
	logger.LogService(func() {
		zlog.Logger.Info().Time("time_now", startTime).Time("expires_at", expiresAt).TimeDiff("lifespan", expiresAt, startTime).Msg("log for booking event")
	})
	if event.CreatedBy == userID {
		return nil, models.ErrBookingOwnEvent
	}
	if event.AvailableSeats < req.SeatCount {
		return nil, models.ErrNotEnoughAvailableSeats
	}

	booking := models.BookingInformation{
		EventID:     req.EventID,
		UserID:      userID,
		SeatCount:   req.SeatCount,
		Status:      models.StatusPending,
		BookingCode: uuid.NewString(),
		ExpiresAt:   expiresAt,
	}

	if err = servs.repo.CreateBookingWithSeatUpdate(ctx, &booking); err != nil {
		logger.LogService(func() { zlog.Logger.Error().Err(err).Msg("failed to create booking") })
		return nil, err
	}

	return &booking, nil
}

func (servs *EventBookerService) ConfirmBooking(ctx context.Context, req *models.ConfirmBookingRequest, userID int) (*models.BookingInformation, error) {
	var err error
	var booking *models.BookingInformation
	if _, err = servs.CheckUserExistsByID(ctx, userID); err != nil {
		return nil, err
	}

	if booking, err = servs.repo.GetBookingByCode(ctx, req.BookingCode); err != nil {
		logger.LogService(func() { zlog.Logger.Error().Err(err).Msg("failed to get booking by code") })
		return nil, err
	}

	if err = servs.repo.ConfirmBooking(ctx, req.BookingCode); err != nil {
		logger.LogService(func() { zlog.Logger.Error().Err(err).Msg("failed to get event by id") })
		return nil, err
	}
	booking.Status = models.StatusConfirmed
	return booking, nil
}

func (servs *EventBookerService) CheckUserExistsByID(ctx context.Context, userID int) (*models.UserInformation, error) {
	user, err := servs.repo.GetUserByID(ctx, userID)
	if err == models.ErrUserNotFound {
		logger.LogService(func() { zlog.Logger.Error().Err(err).Msg("user not found") })
		return nil, err
	}
	if err != nil {
		logger.LogService(func() { zlog.Logger.Error().Err(err).Msg("failed to check if user exists") })
		return nil, err
	}

	return user, nil
}

func (servs *EventBookerService) CheckUserExistsByEmail(ctx context.Context, email string) (*models.UserInformation, error) {
	user, err := servs.repo.GetUserByEmail(ctx, email)
	if err == models.ErrUserNotFound {
		logger.LogService(func() { zlog.Logger.Error().Err(err).Msg("user not found") })
		return nil, err
	}
	if err != nil {
		logger.LogService(func() { zlog.Logger.Error().Err(err).Msg("failed to check if user exists") })
		return nil, err
	}

	return user, nil
}

func hashPassword(password string) (string, error) {
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
func checkPassword(password, hashedPassword string) bool {
	if password == "" || hashedPassword == "" {
		return false
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
