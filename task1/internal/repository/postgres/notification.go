package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/pozedorum/WB_project_3/task1/internal/models"
	"github.com/pozedorum/WB_project_3/task1/internal/service"
	"github.com/pozedorum/wbf/dbpg"
	"github.com/pozedorum/wbf/zlog"
)

type NotificationRepository struct {
	db *dbpg.DB
}

func NewNotificationRepositoryWithDB(masterDSN string, slaveDSNs []string, opts *dbpg.Options) (*NotificationRepository, error) {
	db, err := dbpg.New(masterDSN, slaveDSNs, opts)
	if err != nil {
		return nil, err
	}
	return NewNotificationRepository(db), nil
}

func NewNotificationRepository(db *dbpg.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (nr *NotificationRepository) CreateNotification(ctx context.Context, n *models.Notification) error {
	createQuery := `INSERT INTO notifications (id, user_id, message, channel, send_at, status, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := nr.db.ExecWithRetry(ctx, models.StandartStrategy, createQuery,
		n.ID, n.UserID, n.Message, n.Channel, n.SendAt, n.Status, n.CreatedAt, n.UpdatedAt)

	if err != nil {
		zlog.Logger.Error().Err(err).Str("notification_id", n.ID).Msg("Failed to create notification in database")
	} else {
		zlog.Logger.Info().Str("notification_id", n.ID).Msg("Notification created in database")
	}

	return err
}

func (nr *NotificationRepository) UpdateNotificationStatus(ctx context.Context, id, status string) error {
	updateQuery := `UPDATE notifications SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := nr.db.ExecWithRetry(ctx, models.StandartStrategy, updateQuery, status, time.Now(), id)

	if err != nil {
		zlog.Logger.Error().Err(err).Str("notification_id", id).Str("status", status).Msg("Failed to update notification status")
	} else {
		zlog.Logger.Info().Str("notification_id", id).Str("status", status).Msg("Notification status updated")
	}

	return err
}

func (nr *NotificationRepository) DeleteNotification(ctx context.Context, id string) error {
	deleteQuery := `DELETE FROM notifications WHERE id = $1`
	_, err := nr.db.ExecWithRetry(ctx, models.StandartStrategy, deleteQuery, id)

	if err != nil {
		zlog.Logger.Error().Err(err).Str("notification_id", id).Msg("Failed to delete notification")
	} else {
		zlog.Logger.Info().Str("notification_id", id).Msg("Notification deleted from database")
	}

	return err
}

func (nr *NotificationRepository) GetByID(ctx context.Context, id string) (*models.Notification, error) {
	getQuery := `SELECT id, user_id, message, channel, send_at, status, created_at, updated_at 
		FROM notifications WHERE id = $1`
	var res models.Notification
	rows, err := nr.db.QueryWithRetry(ctx, models.StandartStrategy, getQuery, id)
	if err != nil {
		zlog.Logger.Error().Err(err).Str("notification_id", id).Msg("Query failed for notification")
		return nil, fmt.Errorf("query failed: %v", err)
	}

	if !rows.Next() {
		if err = rows.Err(); err != nil {
			zlog.Logger.Error().Err(err).Str("notification_id", id).Msg("Rows error for notification")
			return nil, fmt.Errorf("rows error: %w", err)
		}
		zlog.Logger.Debug().Str("notification_id", id).Msg("Notification not found in database")
		return nil, nil
	}

	err = rows.Scan(&res.ID, &res.UserID, &res.Message, &res.Channel,
		&res.SendAt, &res.Status, &res.CreatedAt, &res.UpdatedAt)
	if err != nil {
		zlog.Logger.Error().Err(err).Str("notification_id", id).Msg("Scan failed for notification")
		return nil, fmt.Errorf("scan failed: %w", err)
	}

	if rows.Next() {
		zlog.Logger.Error().Str("notification_id", id).Msg("Unexpected multiple rows for notification")
		return nil, fmt.Errorf("unexpected multiple rows for id: %s", id)
	}

	zlog.Logger.Debug().Str("notification_id", id).Msg("Notification retrieved from database")
	return &res, nil
}

func (nr *NotificationRepository) Close() {
	nr.db.Master.Close()
	for _, slave := range nr.db.Slaves {
		if slave != nil {
			slave.Close()
		}
	}
	zlog.Logger.Info().Msg("PostgreSQL connections closed")
}

// Implement Repository interface
var _ service.Repository = (*NotificationRepository)(nil)
