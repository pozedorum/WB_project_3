package server

import (
	"strconv"

	"github.com/pozedorum/WB_project_3/task5/internal/models"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
)

func (serv *EventBookerServer) CreateEvent(c *ginext.Context) {
	var req models.EventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.Logger.Error().Err(err).Msg("Failed to bind JSON for create event")
		c.JSON(models.StatusBadRequest, ginext.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	// Получаем userID из контекста (предполагается, что middleware JWT добавило его)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(models.StatusBadRequest, ginext.H{"error": "User not authenticated"})
		return
	}

	userIDInt, ok := userID.(int)
	if !ok {
		c.JSON(models.StatusBadRequest, ginext.H{"error": "Invalid user ID format"})
		return
	}

	// Создаем событие
	event, err := serv.service.CreateEvent(c.Request.Context(), &req, userIDInt)
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("Failed to create event")
		c.JSON(models.StatusInternalServerError, ginext.H{"error": "Failed to create event"})
		return
	}

	c.JSON(models.StatusOK, ginext.H{
		"message": "Event created successfully",
		"event":   event,
	})
}

func (serv *EventBookerServer) BookEvent(c *ginext.Context) {
	var req models.BookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.Logger.Error().Err(err).Msg("Failed to bind JSON for book event")
		c.JSON(models.StatusBadRequest, ginext.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	// Получаем eventID из URL параметра
	eventIDStr := c.Param("id")
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		c.JSON(models.StatusBadRequest, ginext.H{"error": "Invalid event ID"})
		return
	}

	// Устанавливаем eventID из URL, а не из тела запроса (для безопасности)
	req.EventID = eventID

	// Получаем userID из контекста
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(models.StatusBadRequest, ginext.H{"error": "User not authenticated"})
		return
	}

	userIDInt, ok := userID.(int)
	if !ok {
		c.JSON(models.StatusBadRequest, ginext.H{"error": "Invalid user ID format"})
		return
	}

	// Создаем бронирование
	booking, err := serv.service.BookEvent(c.Request.Context(), &req, userIDInt)
	if err != nil {
		switch err {
		case models.ErrNotEnoughAvailableSeats:
			c.JSON(models.StatusBadRequest, ginext.H{"error": "Not enough available seats"})
		case models.ErrEventNotFound:
			c.JSON(models.StatusNotFound, ginext.H{"error": "Event not found"})
		default:
			zlog.Logger.Error().Err(err).Msg("Failed to book event")
			c.JSON(models.StatusInternalServerError, ginext.H{"error": "Failed to book event"})
		}
		return
	}

	c.JSON(models.StatusOK, ginext.H{
		"message": "Booking created successfully",
		"booking": booking,
	})
}

func (serv *EventBookerServer) ConfirmBooking(c *ginext.Context) {
	var req models.ConfirmBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.Logger.Error().Err(err).Msg("Failed to bind JSON for confirm booking")
		c.JSON(models.StatusBadRequest, ginext.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	// Получаем userID из контекста
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(models.StatusBadRequest, ginext.H{"error": "User not authenticated"})
		return
	}

	userIDInt, ok := userID.(int)
	if !ok {
		c.JSON(models.StatusBadRequest, ginext.H{"error": "Invalid user ID format"})
		return
	}

	// Подтверждаем бронирование
	booking, err := serv.service.ConfirmBooking(c.Request.Context(), &req, userIDInt)
	if err != nil {
		switch err {
		case models.ErrBookingNotFound:
			c.JSON(models.StatusNotFound, ginext.H{"error": "Booking not found"})
		default:
			zlog.Logger.Error().Err(err).Msg("Failed to confirm booking")
			c.JSON(models.StatusInternalServerError, ginext.H{"error": "Failed to confirm booking"})
		}
		return
	}

	c.JSON(models.StatusOK, ginext.H{
		"message": "Booking confirmed successfully",
		"booking": booking,
	})
}

func (serv *EventBookerServer) GetEventInformation(c *ginext.Context) {
	eventIDStr := c.Param("id")

	event, err := serv.service.GetEventInformation(c.Request.Context(), eventIDStr)
	if err != nil {
		switch err {
		case models.ErrEventNotFound:
			c.JSON(models.StatusNotFound, ginext.H{"error": "Event not found"})
		default:
			zlog.Logger.Error().Err(err).Msg("Failed to get event information")
			c.JSON(models.StatusInternalServerError, ginext.H{"error": "Failed to get event information"})
		}
		return
	}

	c.JSON(models.StatusOK, ginext.H{
		"event": event,
	})
}

func (serv *EventBookerServer) GetAllEvents(c *ginext.Context) {
	events, err := serv.service.GetAllEvents(c.Request.Context())
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("Failed to get all events")
		c.JSON(models.StatusInternalServerError, ginext.H{"error": "Failed to get events"})
		return
	}

	c.JSON(models.StatusOK, ginext.H{
		"events": events,
		"count":  len(events),
	})
}

func (serv *EventBookerServer) RegisterUser(c *ginext.Context) {
	var req models.UserRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.Logger.Error().Err(err).Msg("Failed to bind JSON for user registration")
		c.JSON(models.StatusBadRequest, ginext.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	user, err := serv.service.RegisterUser(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case models.ErrUserAlreadyRegistered:
			c.JSON(models.StatusBadRequest, ginext.H{"error": "User already registered"})
		default:
			zlog.Logger.Error().Err(err).Msg("Failed to register user")
			c.JSON(models.StatusInternalServerError, ginext.H{"error": "Failed to register user"})
		}
		return
	}

	c.JSON(models.StatusOK, ginext.H{
		"message": "User registered successfully",
		"user":    user,
	})
}

func (serv *EventBookerServer) LoginUser(c *ginext.Context) {
	var req models.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zlog.Logger.Error().Err(err).Msg("Failed to bind JSON for user login")
		c.JSON(models.StatusBadRequest, ginext.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	user, err := serv.service.AuthUser(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case models.ErrUserNotFound, models.ErrWrongPassword:
			c.JSON(models.StatusBadRequest, ginext.H{"error": "Invalid email or password"})
		default:
			zlog.Logger.Error().Err(err).Msg("Failed to authenticate user")
			c.JSON(models.StatusInternalServerError, ginext.H{"error": "Failed to authenticate user"})
		}
		return
	}

	// Генерируем JWT токен
	token, expiresAt, err := serv.generateJWTToken(user.ID)
	if err != nil {
		zlog.Logger.Error().Err(err).Int("user_id", user.ID).Msg("Failed to generate JWT token")
		c.JSON(models.StatusInternalServerError, ginext.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(models.StatusOK, ginext.H{
		"message":    "Login successful",
		"user":       user,
		"token":      token,
		"expires_at": expiresAt,
	})
}
