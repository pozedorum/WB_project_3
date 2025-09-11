package repository

import (
	"context"
	"database/sql"
	"fmt"

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

// CheckUserExistsByEmail - проверка по email (для регистрации)
func (repo *EventBookerRepository) CheckUserExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	var exists bool
	err := repo.db.Master.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		zlog.Logger.Error().Err(err).Str("email", email).Msg("Failed to check user existence by email")
		return false, err
	}
	return exists, nil
}

// CheckUserExistsByID - проверка по ID (для операций с существующими пользователями)
func (repo *EventBookerRepository) CheckUserExistsByID(ctx context.Context, userID int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`
	var exists bool
	err := repo.db.Master.QueryRowContext(ctx, query, userID).Scan(&exists)
	if err != nil {
		zlog.Logger.Error().Err(err).Int("user_id", userID).Msg("Failed to check user existence by ID")
		return false, err
	}
	return exists, nil
}

// GetUserByID - получение пользователя по ID (если нужна полная информация)
func (repo *EventBookerRepository) GetUserByID(ctx context.Context, userID int) (*models.UserInformation, error) {
	query := `SELECT id, email, name, phone, email_verified, created_at, updated_at 
              FROM users WHERE id = $1`

	var user models.UserInformation
	err := repo.db.Master.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Phone,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		zlog.Logger.Error().Err(err).Int("user_id", userID).Msg("Failed to get user by ID")
		return nil, err
	}

	return &user, nil
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

func (repo *EventBookerRepository) GetEventByID(ctx context.Context, id int) (*models.EventInformation, error) {
	selectQuery := `SELECT name,date,cost, total_seats, available_seats,booking_lifespan_minutes, created_by, created_at, updated_at
	FROM events WHERE id = $1`
	var res models.EventInformation
	err := repo.db.Master.QueryRowContext(ctx, selectQuery, id).Scan(&res.Name, &res.Date, &res.Cost,
		&res.TotalSeats, &res.AvailableSeats, &res.LifeSpan, &res.CreatedBy, &res.CreatedAt, &res.UpdatedAt)
	if err != nil {
		zlog.Logger.Error().Err(err).Int("id", id).Msg("Failed to get event")
		return nil, err
	}
	return &res, nil
}

func (repo *EventBookerRepository) CreateBookingWithSeatUpdate(ctx context.Context, booking *models.BookingInformation) error {
	return repo.WithTransaction(ctx, func(tx *sql.Tx) error {
		// 1. Проверяем и блокируем доступные места
		var availableSeats int
		selectQuery := `SELECT available_seats FROM events WHERE id = $1 FOR UPDATE`
		err := tx.QueryRowContext(ctx, selectQuery, booking.EventID).Scan(&availableSeats)
		if err != nil {
			zlog.Logger.Error().Err(err).Int("event_id", booking.EventID).Msg("Failed to get event seats")
			return fmt.Errorf("event not found")
		}

		// 2. Проверяем достаточно ли мест
		if availableSeats < booking.SeatCount {
			zlog.Logger.Warn().Int("event_id", booking.EventID).
				Int("requested", booking.SeatCount).
				Int("available", availableSeats).
				Msg("Not enough seats available")
			return fmt.Errorf("not enough seats available")
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
				return fmt.Errorf("booking not found")
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

	zlog.Logger.Info().Int("count", len(bookings)).Msg("Found expired bookings")
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

func (repo *EventBookerRepository) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	var exists bool
	err := repo.db.Master.QueryRowContext(ctx, query, email).Scan(&exists)
	return exists, err
}

//var _ service.Repository = (*EventBookerRepository)(nil)
