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
		zlog.Logger.Info().Msg("Cron worker stopped")
	}()
}

func (servs *EventBookerService) processExpiredBookings(ctx context.Context) {
	// Проверяем контекст в начале
	select {
	case <-ctx.Done():
		return
	default:
	}

	expiredBookings, err := servs.repo.GetExpiredBookings(ctx)
	if err != nil {
		// Проверяем, не была ли ошибка вызвана отменой контекста
		if ctx.Err() != nil {
			return
		}
		logger.LogService(func() {
			zlog.Logger.Error().Err(err).Msg("Failed to get expired bookings")
		})
		return
	}

	for _, booking := range expiredBookings {
		// Проверяем контекст перед каждой итерацией
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := servs.repo.CancelBooking(ctx, booking.ID); err != nil {
			if ctx.Err() != nil {
				return
			}
			logger.LogService(func() {
				zlog.Logger.Error().
					Err(err).
					Int("booking_id", booking.ID).
					Msg("Failed to cancel expired booking")
			})
			continue
		}
	}
}
