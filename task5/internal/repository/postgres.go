package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/pozedorum/WB_project_3/task5/internal/models"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/zlog"
)

type EventBookerRepository struct {
	db *dbpg.DB
}

func NewEventBookerRepositoryWithDB(masterDSN string, slaveDSNs []string, opts *dbpg.Options) (*EventBookerRepository, error) {
	db, err := dbpg.New(masterDSN, slaveDSNs, opts)
	if err != nil {
		return nil, err
	}
	return NewEventBookerRepository(db), nil
}

func NewEventBookerRepository(db *dbpg.DB) *EventBookerRepository {
	return &EventBookerRepository{db: db}
}

func (repo *EventBookerRepository) Close() {
	if err := repo.db.Master.Close(); err != nil {
		zlog.Logger.Panic().Msg("Database failed to close")
	}
	for _, slave := range repo.db.Slaves {
		if slave != nil {
			if err := slave.Close(); err != nil {
				zlog.Logger.Panic().Msg("Slave database failed to close")
			}
		}
	}
	zlog.Logger.Info().Msg("PostgreSQL connections closed")
}

// WithTransaction выполняет функцию в транзакции
func (repo *EventBookerRepository) WithTransaction(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := repo.db.Master.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // Safe rollback if not committed

	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		zlog.Logger.Error().Err(err).Msg("Failed to commit transaction")
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (repo *EventBookerRepository) GetUserByID(ctx context.Context, userID int) (*models.UserInformation, error) {
	query := `SELECT id, email, password_hash, name, phone, created_at, updated_at 
              FROM users WHERE id = $1`

	var user models.UserInformation
	err := repo.db.Master.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Name,
		&user.Phone,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrUserNotFound
		}
		zlog.Logger.Error().Err(err).Int("user_id", userID).Msg("Failed to get user by ID")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (repo *EventBookerRepository) GetUserByEmail(ctx context.Context, email string) (*models.UserInformation, error) {
	query := `SELECT id, email, password_hash, name, phone, created_at, updated_at 
              FROM users WHERE email = $1`

	var user models.UserInformation
	err := repo.db.Master.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Name,
		&user.Phone,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrUserNotFound
		}
		zlog.Logger.Error().Err(err).Str("email", email).Msg("Failed to get user by email")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (repo *EventBookerRepository) GetUserHash(ctx context.Context, email string) (string, error) {
	selectQuery := `SELECT password_hash FROM users WHERE email = $1`
	var userHash string
	tx, err := repo.db.Master.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	err = tx.QueryRowContext(ctx, selectQuery, email).Scan(&userHash)
	if err != nil {
		return "", err
	}

	if err := tx.Commit(); err != nil {
		zlog.Logger.Error().Err(err).Msg("Failed to commit transaction")
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}
	return userHash, nil
}

func (repo *EventBookerRepository) GetEventByID(ctx context.Context, id int) (*models.EventInformation, error) {
	selectQuery := `SELECT name, date, cost, total_seats, available_seats, 
                    booking_lifespan_minutes, created_by, created_at, updated_at
                    FROM events WHERE id = $1`

	var res models.EventInformation
	var lifespanMinutes int // ← Храним минуты как число

	err := repo.db.Master.QueryRowContext(ctx, selectQuery, id).Scan(
		&res.Name, &res.Date, &res.Cost,
		&res.TotalSeats, &res.AvailableSeats, &lifespanMinutes, // ← Сканируем в минуты
		&res.CreatedBy, &res.CreatedAt, &res.UpdatedAt)

	if err != nil {
		zlog.Logger.Error().Err(err).Int("id", id).Msg("Failed to get event")
		return nil, err
	}

	// Конвертируем минуты обратно в Duration
	res.LifeSpan = time.Duration(lifespanMinutes) * time.Minute
	zlog.Logger.Info().Int("lifespan_minutes_from_table", lifespanMinutes).Dur("lifespan_duration", res.LifeSpan).Msg("getEventByID")
	return &res, nil
}

func (repo *EventBookerRepository) GetAllEvents(ctx context.Context) ([]*models.EventInformation, error) {
	query := `SELECT id, name, date, cost, total_seats, available_seats, 
                     booking_lifespan_minutes, created_by, created_at, updated_at
              FROM events 
              ORDER BY created_at DESC`

	var events []*models.EventInformation
	rows, err := repo.db.Master.QueryContext(ctx, query)
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("Failed to get all events")
		return nil, fmt.Errorf("failed to get events: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var event models.EventInformation
		var lifespanMinutes int // ← Добавляем временную переменную

		err := rows.Scan(
			&event.ID,
			&event.Name,
			&event.Date,
			&event.Cost,
			&event.TotalSeats,
			&event.AvailableSeats,
			&lifespanMinutes, // ← Сканируем минуты
			&event.CreatedBy,
			&event.CreatedAt,
			&event.UpdatedAt,
		)
		if err != nil {
			zlog.Logger.Error().Err(err).Msg("Failed to scan event")
			continue
		}

		// Конвертируем обратно в Duration
		event.LifeSpan = time.Duration(lifespanMinutes) * time.Minute

		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		zlog.Logger.Error().Err(err).Msg("Error iterating events")
		return nil, fmt.Errorf("error iterating events: %w", err)
	}

	return events, nil
}

func (repo *EventBookerRepository) GetBookingByCode(ctx context.Context, bookingCode string) (*models.BookingInformation, error) {
	selectQuery := `SELECT id, event_id, user_id, seat_count, status, booking_code,
                           created_at, expires_at, confirmed_at
                    FROM bookings 
                    WHERE booking_code = $1`
	var booking models.BookingInformation
	err := repo.db.Master.QueryRowContext(ctx, selectQuery, bookingCode).Scan(
		&booking.ID, &booking.EventID, &booking.UserID, &booking.SeatCount,
		&booking.Status, &booking.BookingCode, &booking.CreatedAt,
		&booking.ExpiresAt, &booking.ConfirmedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrBookingNotFound // ← ДОБАВИТЬ обработку "не найдено"
		}
		zlog.Logger.Error().Err(err).Msg("Failed to get booking")
		return nil, err
	}
	return &booking, nil

}

func (repo *EventBookerRepository) CreateUser(ctx context.Context, n *models.UserInformation) error {
	return repo.WithTransaction(ctx, func(tx *sql.Tx) error {
		createQuery := `INSERT INTO users (email, password_hash, name, phone)
                        VALUES($1, $2, $3, $4) 
                        RETURNING id, created_at, updated_at`
		err := tx.QueryRowContext(ctx, createQuery,
			n.Email,
			n.PasswordHash,
			n.Name,
			n.Phone).
			Scan(&n.ID, &n.CreatedAt, &n.UpdatedAt)
		if err != nil {
			zlog.Logger.Error().Err(err).Str("email", n.Email).Msg("Failed to create user")
			return err
		}

		zlog.Logger.Info().Str("email", n.Email).Int("user_id", n.ID).Msg("User created successfully")
		return nil
	})
}

func (repo *EventBookerRepository) CreateEvent(ctx context.Context, n *models.EventInformation) error {
	return repo.WithTransaction(ctx, func(tx *sql.Tx) error {
		createQuery := `INSERT INTO events (name, date, cost, total_seats, available_seats, 
										booking_lifespan_minutes, created_by)
					VALUES($1,$2,$3,$4,$5,$6, $7) RETURNING id, created_at, updated_at`

		lifespanMinutes := int(n.LifeSpan.Minutes())

		err := tx.QueryRowContext(ctx, createQuery,
			n.Name,
			n.Date,
			n.Cost,
			n.TotalSeats,
			n.TotalSeats,
			lifespanMinutes,
			n.CreatedBy,
		).Scan(&n.ID, &n.CreatedAt, &n.UpdatedAt)

		if err != nil {
			zlog.Logger.Error().Err(err).Str("event_name", n.Name).Msg("Failed to create event in database")
			return err
		}

		zlog.Logger.Info().Int("id", n.ID).Str("event_name", n.Name).Msg("Event created in database")
		return nil
	})
}

func (repo *EventBookerRepository) CreateBookingWithSeatUpdate(ctx context.Context, booking *models.BookingInformation) error {
	return repo.WithTransaction(ctx, func(tx *sql.Tx) error {
		// 1. Проверяем и блокируем доступные места
		var availableSeats int
		selectQuery := `SELECT available_seats FROM events WHERE id = $1 FOR UPDATE`
		err := tx.QueryRowContext(ctx, selectQuery, booking.EventID).Scan(&availableSeats)
		if err != nil {
			zlog.Logger.Error().Err(err).Int("event_id", booking.EventID).Msg("Failed to get event seats")
			return models.ErrEventNotFound
		}

		// 2. Проверяем достаточно ли мест
		if availableSeats < booking.SeatCount {
			zlog.Logger.Warn().Int("event_id", booking.EventID).
				Int("requested", booking.SeatCount).
				Int("available", availableSeats).
				Msg("Not enough seats available")
			return models.ErrNotEnoughAvailableSeats
		}

		// 3. Уменьшаем количество свободных мест
		updateSeatsQuery := `UPDATE events SET available_seats = available_seats - $1, 
                            updated_at = NOW() WHERE id = $2`
		_, err = tx.ExecContext(ctx, updateSeatsQuery, booking.SeatCount, booking.EventID)
		if err != nil {
			zlog.Logger.Error().Err(err).Int("event_id", booking.EventID).
				Int("seat_count", booking.SeatCount).
				Msg("Failed to update available seats")
			return err
		}

		// 4. Создаем бронирование
		insertQuery := `INSERT INTO bookings (event_id, user_id, seat_count, status, 
                        booking_code, expires_at)
                        VALUES ($1, $2, $3, $4, $5, $6) 
                        RETURNING id, created_at`

		err = tx.QueryRowContext(ctx, insertQuery,
			booking.EventID,
			booking.UserID,
			booking.SeatCount,
			booking.Status,
			booking.BookingCode,
			booking.ExpiresAt,
		).Scan(&booking.ID, &booking.CreatedAt)

		if err != nil {
			zlog.Logger.Error().Err(err).
				Int("user_id", booking.UserID).
				Int("event_id", booking.EventID).
				Msg("Failed to create booking")
			return err
		}

		zlog.Logger.Info().
			Int("id", booking.ID).
			Int("user_id", booking.UserID).
			Int("event_id", booking.EventID).
			Msg("Booking created successfully with seat update")

		return nil
	})
}

func (repo *EventBookerRepository) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	var exists bool
	err := repo.db.Master.QueryRowContext(ctx, query, email).Scan(&exists)
	return exists, err
}

func (repo *EventBookerRepository) ConfirmBooking(ctx context.Context, bookingCode string) error {
	return repo.WithTransaction(ctx, func(tx *sql.Tx) error {
		// 1. Находим бронь и блокируем ее
		var bookingID, eventID, seatCount int
		selectQuery := `SELECT id, event_id, seat_count FROM bookings 
                       WHERE booking_code = $1 AND status = 'pending' FOR UPDATE`

		err := tx.QueryRowContext(ctx, selectQuery, bookingCode).Scan(&bookingID, &eventID, &seatCount)
		if err != nil {
			if err == sql.ErrNoRows {
				zlog.Logger.Warn().Str("booking_code", bookingCode).Msg("Booking not found or already processed")
				return models.ErrBookingNotFound
			}
			zlog.Logger.Error().Err(err).Str("booking_code", bookingCode).Msg("Failed to find booking")
			return err
		}

		// 2. Подтверждаем бронь
		updateBookingQuery := `UPDATE bookings SET status = 'confirmed', confirmed_at = NOW() 
                              WHERE id = $1`
		_, err = tx.ExecContext(ctx, updateBookingQuery, bookingID)
		if err != nil {
			zlog.Logger.Error().Err(err).Int("booking_id", bookingID).Msg("Failed to confirm booking")
			return err
		}

		// 3. Места уже были заняты при создании брони, поэтому не обновляем available_seats
		zlog.Logger.Info().Str("booking_code", bookingCode).Int("booking_id", bookingID).Msg("Booking confirmed successfully")
		return nil
	})
}

func (repo *EventBookerRepository) GetExpiredBookings(ctx context.Context) ([]*models.BookingInformation, error) {
	selectQuery := `SELECT id, event_id, user_id, seat_count, status, booking_code,
                           created_at, expires_at, confirmed_at
                    FROM bookings 
                    WHERE status = 'pending' AND expires_at < NOW()`

	var bookings []*models.BookingInformation
	rows, err := repo.db.Master.QueryContext(ctx, selectQuery)
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("Failed to get expired bookings")
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var booking models.BookingInformation
		err := rows.Scan(
			&booking.ID, &booking.EventID, &booking.UserID, &booking.SeatCount,
			&booking.Status, &booking.BookingCode, &booking.CreatedAt,
			&booking.ExpiresAt, &booking.ConfirmedAt,
		)
		if err != nil {
			zlog.Logger.Error().Err(err).Msg("Failed to scan expired booking")
			continue
		}
		bookings = append(bookings, &booking)
	}

	if err := rows.Err(); err != nil {
		zlog.Logger.Error().Err(err).Msg("Error iterating expired bookings")
		return nil, err
	}
	if len(bookings) != 0 {
		zlog.Logger.Info().Int("count", len(bookings)).Msg("Found expired bookings")
	} else {
		zlog.Logger.Info().Msg("No bookings expired")
	}

	return bookings, nil
}

func (repo *EventBookerRepository) CancelBooking(ctx context.Context, bookingID int) error {
	return repo.WithTransaction(ctx, func(tx *sql.Tx) error {
		// 1. Получаем информацию о брони для освобождения мест
		var eventID, seatCount int
		var status string

		selectQuery := `SELECT event_id, seat_count, status FROM bookings WHERE id = $1 FOR UPDATE`
		err := tx.QueryRowContext(ctx, selectQuery, bookingID).Scan(&eventID, &seatCount, &status)
		if err != nil {
			zlog.Logger.Error().Err(err).Int("booking_id", bookingID).Msg("Failed to find booking")
			return err
		}

		// 2. Отменяем бронь
		updateBookingQuery := `UPDATE bookings SET status = 'cancelled' WHERE id = $1`
		_, err = tx.ExecContext(ctx, updateBookingQuery, bookingID)
		if err != nil {
			zlog.Logger.Error().Err(err).Int("booking_id", bookingID).Msg("Failed to cancel booking")
			return err
		}

		// 3. Освобождаем места только если бронь была активной
		if status == "pending" || status == "confirmed" {
			updateSeatsQuery := `UPDATE events SET available_seats = available_seats + $1, 
                                updated_at = NOW() WHERE id = $2`
			_, err = tx.ExecContext(ctx, updateSeatsQuery, seatCount, eventID)
			if err != nil {
				zlog.Logger.Error().Err(err).Int("event_id", eventID).Int("seats", seatCount).Msg("Failed to free seats")
				return err
			}
		}

		zlog.Logger.Info().Int("booking_id", bookingID).Int("event_id", eventID).Msg("Booking cancelled successfully")
		return nil
	})
}

// var _ service.Repository = (*EventBookerRepository)(nil)
