package service

import (
	"context"

	"github.com/pozedorum/WB_project_3/task5/pkg/logger"
	"github.com/robfig/cron/v3"
	"github.com/wb-go/wbf/zlog"
)

// Методы для обработки просроченных бронирований
func (servs *EventBookerService) StartCronWorker(ctx context.Context) {
	c := cron.New()

	// Запускаем каждую минуту
	c.AddFunc("* * * * *", func() {
		servs.processExpiredBookings(ctx)
	})

	c.Start()

	// Останавливаем при завершении контекста
	go func() {
		<-ctx.Done()
		c.Stop()
	}()
}

func (servs *EventBookerService) processExpiredBookings(ctx context.Context) {
	expiredBookings, err := servs.repo.GetExpiredBookings(ctx)
	if err != nil {
		logger.LogService(func() {
			zlog.Logger.Error().Err(err).Msg("Failed to get expired bookings")
		})
		return
	}

	for _, booking := range expiredBookings {
		if err := servs.repo.CancelBooking(ctx, booking.ID); err != nil {
			logger.LogService(func() {
				zlog.Logger.Error().
					Err(err).
					Int("booking_id", booking.ID).
					Msg("Failed to cancel expired booking")
			})
			continue
		}
		logger.LogService(func() {
			zlog.Logger.Info().
				Int("booking_id", booking.ID).
				Str("booking_code", booking.BookingCode).
				Msg("Successfully cancelled expired booking")
		})

		// ДОПОЛНИТЕЛЬНО: можно добавить отправку уведомления пользователю
	}
}
