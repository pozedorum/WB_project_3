package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/pozedorum/WB_project_3/task1/internal/models"
	"github.com/pozedorum/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
)

var standartStrategy = retry.Strategy{Attempts: 3, Delay: time.Second}

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
	createQuery := `INSERT INTO Notifications (id, user_id, message, channel, send_at, status, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := nr.db.ExecWithRetry(ctx, standartStrategy, createQuery,
		n.ID, n.UserID, n.Message, n.Channel, n.SendAt, n.Status, n.CreatedAt, n.UpdatedAt)

	return err
}

func (nr *NotificationRepository) UpdateNotificationStatus(ctx context.Context, id, status string) error {
	updateQuery := `UPDATE Notifications SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := nr.db.ExecWithRetry(ctx, standartStrategy, updateQuery, status, time.Now(), id)

	return err
}

func (nr *NotificationRepository) DeleteNotification(ctx context.Context, id string) error {
	deleteQuery := `DELETE FROM Notifications WHERE id = $1`
	_, err := nr.db.ExecWithRetry(ctx, standartStrategy, deleteQuery, id)
	return err
}

func (nr *NotificationRepository) GetByID(ctx context.Context, id string) (*models.Notification, error) {
	getQuery := `SELECT id, user_id, message, channel, send_at, status, created_at, updated_at 
		FROM Notifications WHERE id = $1`
	var res models.Notification
	rows, err := nr.db.QueryWithRetry(ctx, standartStrategy, getQuery, id)
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}

	if !rows.Next() {
		if err = rows.Err(); err != nil {
			return nil, fmt.Errorf("rows error: %w", err)
		}
		return nil, nil
	}

	err = rows.Scan(&res.ID, &res.UserID, &res.Message, &res.Channel,
		&res.SendAt, &res.Status, &res.CreatedAt, &res.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("scan failed: %w", err)
	}

	if rows.Next() {
		return nil, fmt.Errorf("unexpected multiple rows for id: %s", id)
	}
	return &res, nil
}
