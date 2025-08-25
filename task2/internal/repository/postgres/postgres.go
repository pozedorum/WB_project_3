package postgres

import (
	"context"
	"fmt"

	"github.com/pozedorum/WB_project_3/task2/internal/models"
	"github.com/pozedorum/wbf/dbpg"
	"github.com/pozedorum/wbf/zlog"
)

type ShortURLRepository struct {
	db *dbpg.DB
}

func NewShortURLRepositoryWithDB(masterDSN string, slaveDSNs []string, opts *dbpg.Options) (*ShortURLRepository, error) {
	db, err := dbpg.New(masterDSN, slaveDSNs, opts)
	if err != nil {
		return nil, err
	}
	return NewShortURLRepository(db), nil
}

func NewShortURLRepository(db *dbpg.DB) *ShortURLRepository {
	return &ShortURLRepository{db: db}
}

func (nr *ShortURLRepository) CreateShortURL(ctx context.Context, n *models.ShortURL) error {
	createQuery := `INSERT INTO short_urls (id, short_code, original_url, created_at, expires_at, is_active, clicks_count) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := nr.db.ExecWithRetry(ctx, models.StandartStrategy, createQuery,
		n.ID, n.ShortCode, n.OriginalURL, n.CreatedAt, n.ExpiresAt, n.IsActive, n.Clicks)

	if err != nil {
		zlog.Logger.Error().Err(err).Str("url_id", n.ID).Msg("Failed to create url in database")
	} else {
		zlog.Logger.Info().Str("url_id", n.ID).Msg("URL created in database")
	}

	return err
}

func (nr *ShortURLRepository) RegisterClick(ctx context.Context, n *models.ClickAnalyticsEntry) error {
	getIDQuery := `SELECT id FROM short_urls WHERE short_code = $1`
	createQuery := `INSERT INTO short_urls (id, short_code, user_agent, ip_address, referrer, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6)`
	updateQuery := `UPDATE short_urls 
        SET clicks_count = clicks_count + 1 
        WHERE short_code = $1`

	tx, err := nr.db.Master.Begin()
	if err != nil {
		zlog.Logger.Error().Err(err).Str("url_id", n.ID).Msg("Failed to start transaction")
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, getIDQuery, n.ShortURLID)
	_, err = tx.ExecContext(ctx, createQuery, n.ID, n.ShortURLID, n.UserAgent, n.IPAddress, n.Referrer, n.CreatedAt)
	_, err = tx.ExecContext(ctx, updateQuery, n.ShortURLID)
	if err != nil {
		zlog.Logger.Error().Err(err).Str("url_id", n.ID).Msg("Failed to create url in database")
	} else {
		zlog.Logger.Info().Str("url_id", n.ID).Msg("URL created in database")
	}

	return err
}

func (nr *ShortURLRepository) DeleteURL(ctx context.Context, id string) error {
	deleteQuery := `DELETE FROM urls WHERE id = $1`
	_, err := nr.db.ExecWithRetry(ctx, models.StandartStrategy, deleteQuery, id)

	if err != nil {
		zlog.Logger.Error().Err(err).Str("url_id", id).Msg("Failed to delete url")
	} else {
		zlog.Logger.Info().Str("url_id", id).Msg("URL deleted from database")
	}

	return err
}

func (nr *ShortURLRepository) GetByID(ctx context.Context, id string) (*models.URL, error) {
	getQuery := `SELECT id, user_id, message, channel, send_at, status, created_at, updated_at 
		FROM urls WHERE id = $1`
	var res models.URL
	rows, err := nr.db.QueryWithRetry(ctx, models.StandartStrategy, getQuery, id)
	if err != nil {
		zlog.Logger.Error().Err(err).Str("url_id", id).Msg("Query failed for url")
		return nil, fmt.Errorf("query failed: %v", err)
	}

	if !rows.Next() {
		if err = rows.Err(); err != nil {
			zlog.Logger.Error().Err(err).Str("url_id", id).Msg("Rows error for url")
			return nil, fmt.Errorf("rows error: %w", err)
		}
		zlog.Logger.Debug().Str("url_id", id).Msg("URL not found in database")
		return nil, nil
	}

	err = rows.Scan(&res.ID, &res.UserID, &res.Message, &res.Channel,
		&res.SendAt, &res.Status, &res.CreatedAt, &res.UpdatedAt)
	if err != nil {
		zlog.Logger.Error().Err(err).Str("url_id", id).Msg("Scan failed for url")
		return nil, fmt.Errorf("scan failed: %w", err)
	}

	if rows.Next() {
		zlog.Logger.Error().Str("url_id", id).Msg("Unexpected multiple rows for url")
		return nil, fmt.Errorf("unexpected multiple rows for id: %s", id)
	}

	zlog.Logger.Debug().Str("url_id", id).Msg("URL retrieved from database")
	return &res, nil
}

func (nr *ShortURLRepository) Close() {
	nr.db.Master.Close()
	for _, slave := range nr.db.Slaves {
		if slave != nil {
			slave.Close()
		}
	}
	zlog.Logger.Info().Msg("PostgreSQL connections closed")
}

// Implement Repository interface
var _ service.Repository = (*ShortURLRepository)(nil)
