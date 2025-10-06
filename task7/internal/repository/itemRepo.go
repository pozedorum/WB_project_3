package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pozedorum/WB_project_3/task5/pkg/logger"
	"github.com/pozedorum/WB_project_3/task7/internal/models"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/zlog"
)

type ItemRepository struct {
	db *dbpg.DB
}

func New(db *dbpg.DB) *ItemRepository {
	logger.LogRepository(func() {
		zlog.Logger.Info().Msg("Creating new ItemRepository instance")
	})
	return &ItemRepository{db: db}
}

func (repo *ItemRepository) Close() error {
	logger.LogRepository(func() {
		zlog.Logger.Info().Msg("Closing ItemRepository database connections")
	})
	if err := repo.db.Master.Close(); err != nil {
		logger.LogRepository(func() {
			zlog.Logger.Error().Err(err).Msg("Master database failed to close")
		})
		return err
	}
	for i, slave := range repo.db.Slaves {
		if slave != nil {
			if err := slave.Close(); err != nil {
				logger.LogRepository(func() {
					zlog.Logger.Error().Err(err).Int("slave_index", i).Msg("Slave database failed to close")
				})
				return err
			}
		}
	}
	logger.LogRepository(func() {
		zlog.Logger.Info().Msg("ItemRepository PostgreSQL connections closed successfully")
	})
	return nil
}

func (repo *ItemRepository) Create(ctx context.Context, item *models.Item) error {
	logger.LogRepository(func() {
		zlog.Logger.Info().
			Str("name", item.Name).
			Int64("price", item.Price).
			Str("created_by", item.CreatedBy).
			Msg("Creating new item")
	})

	query := `INSERT INTO items (name, price, created_by) 
	          VALUES ($1, $2, $3) 
	          RETURNING id, created_at, updated_at`

	err := repo.db.Master.QueryRowContext(ctx, query,
		item.Name, item.Price, item.CreatedBy,
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		logger.LogRepository(func() {
			zlog.Logger.Error().
				Err(err).
				Str("name", item.Name).
				Msg("Failed to create item")
		})
		return err
	}

	logger.LogRepository(func() {
		zlog.Logger.Info().
			Int64("id", item.ID).
			Str("name", item.Name).
			Msg("Item created successfully")
	})
	return nil
}

func (repo *ItemRepository) FindByID(ctx context.Context, id int64) (*models.Item, error) {
	logger.LogRepository(func() {
		zlog.Logger.Info().
			Int64("id", id).
			Msg("Finding item by ID")
	})

	query := `SELECT id, name, price, created_by, created_at, updated_at
	          FROM items WHERE id = $1`

	var item models.Item
	err := repo.db.Master.QueryRowContext(ctx, query, id).Scan(
		&item.ID, &item.Name, &item.Price, &item.CreatedBy,
		&item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.LogRepository(func() {
				zlog.Logger.Warn().
					Int64("id", id).
					Msg("Item not found")
			})
			return nil, sql.ErrNoRows
		}
		logger.LogRepository(func() {
			zlog.Logger.Error().
				Err(err).
				Int64("id", id).
				Msg("Error in FindByID query")
		})
		return nil, err
	}

	logger.LogRepository(func() {
		zlog.Logger.Info().
			Int64("id", id).
			Str("name", item.Name).
			Msg("Item found successfully")
	})
	return &item, nil
}

func (repo *ItemRepository) FindAll(ctx context.Context) ([]models.Item, error) {
	logger.LogRepository(func() {
		zlog.Logger.Info().Msg("Finding all items")
	})

	query := `SELECT id, name, price, created_by, created_at, updated_at
	          FROM items ORDER BY created_at DESC`

	rows, err := repo.db.Master.QueryContext(ctx, query)
	if err != nil {
		logger.LogRepository(func() {
			zlog.Logger.Error().
				Err(err).
				Msg("Error in FindAll query")
		})
		return nil, err
	}
	defer rows.Close()

	var items []models.Item
	for rows.Next() {
		var item models.Item
		err := rows.Scan(
			&item.ID, &item.Name, &item.Price, &item.CreatedBy,
			&item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			logger.LogRepository(func() {
				zlog.Logger.Error().
					Err(err).
					Msg("Error scanning item row")
			})
			return nil, err
		}
		items = append(items, item)
	}

	logger.LogRepository(func() {
		zlog.Logger.Info().
			Int("count", len(items)).
			Msg("Items retrieved successfully")
	})
	return items, nil
}

func (repo *ItemRepository) Update(ctx context.Context, id int64, item *models.Item) error {
	logger.LogRepository(func() {
		zlog.Logger.Info().
			Int64("id", id).
			Str("name", item.Name).
			Int64("price", item.Price).
			Str("created_by", item.CreatedBy).
			Msg("Updating item")
	})

	query := `UPDATE items 
	          SET name = $1, price = $2, updated_at = NOW()
	          WHERE id = $3
	          RETURNING updated_at`

	err := repo.db.Master.QueryRowContext(ctx, query,
		item.Name, item.Price, id,
	).Scan(&item.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.LogRepository(func() {
				zlog.Logger.Warn().
					Int64("id", id).
					Msg("No rows affected - item not found")
			})
			return sql.ErrNoRows // Запись не найдена
		}
		logger.LogRepository(func() {
			zlog.Logger.Error().
				Err(err).
				Int64("id", id).
				Msg("Failed to update item")
		})
		return err
	}

	logger.LogRepository(func() {
		zlog.Logger.Info().
			Int64("id", id).
			Time("updated_at", item.UpdatedAt).
			Msg("Item updated successfully")
	})
	return nil
}

func (repo *ItemRepository) Delete(ctx context.Context, id int64) error {
	logger.LogRepository(func() {
		zlog.Logger.Info().
			Int64("id", id).
			Msg("Deleting item")
	})

	query := `DELETE FROM items WHERE id = $1`
	result, err := repo.db.Master.ExecContext(ctx, query, id)
	if err != nil {
		logger.LogRepository(func() {
			zlog.Logger.Error().
				Err(err).
				Int64("id", id).
				Msg("Failed to delete item")
		})
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.LogRepository(func() {
			zlog.Logger.Error().
				Err(err).
				Int64("id", id).
				Msg("Failed to get rows affected for delete")
		})
		return err
	}

	if rowsAffected == 0 {
		logger.LogRepository(func() {
			zlog.Logger.Warn().
				Int64("id", id).
				Msg("No rows affected when deleting item - item not found")
		})
		return sql.ErrNoRows
	}

	logger.LogRepository(func() {
		zlog.Logger.Info().
			Int64("id", id).
			Int64("rows_affected", rowsAffected).
			Msg("Item deleted successfully")
	})
	return nil
}

func (repo *ItemRepository) GetItemHistory(ctx context.Context, itemID int64) ([]models.HistoryRecord, error) {
	logger.LogRepository(func() {
		zlog.Logger.Info().
			Int64("item_id", itemID).
			Msg("Getting item history")
	})

	query := `SELECT id, item_id, action, changed_by, changed_at
	          FROM item_history 
	          WHERE item_id = $1 
	          ORDER BY changed_at DESC`

	rows, err := repo.db.Master.QueryContext(ctx, query, itemID)
	if err != nil {
		logger.LogRepository(func() {
			zlog.Logger.Error().
				Err(err).
				Int64("item_id", itemID).
				Msg("Error getting item history")
		})
		return nil, err
	}
	defer rows.Close()

	var history []models.HistoryRecord
	for rows.Next() {
		var record models.HistoryRecord
		err := rows.Scan(
			&record.ID, &record.ItemID, &record.Action,
			&record.ChangedBy, &record.ChangedAt,
		)
		if err != nil {
			logger.LogRepository(func() {
				zlog.Logger.Error().
					Err(err).
					Int64("item_id", itemID).
					Msg("Error scanning history record")
			})
			return nil, err
		}
		history = append(history, record)
	}

	logger.LogRepository(func() {
		zlog.Logger.Info().
			Int64("item_id", itemID).
			Int("records_count", len(history)).
			Msg("Item history retrieved successfully")
	})
	return history, nil
}

func (repo *ItemRepository) GetAllHistory(ctx context.Context, filters map[string]interface{}) ([]models.HistoryRecord, error) {
	logger.LogRepository(func() {
		zlog.Logger.Info().
			Interface("filters", filters).
			Msg("Getting all history with filters")
	})

	query := `SELECT id, item_id, action, changed_by, changed_at
	          FROM item_history WHERE 1=1`
	args := []interface{}{}
	argIndex := 1

	if changedBy, ok := filters["changed_by"].(string); ok && changedBy != "" {
		query += fmt.Sprintf(" AND changed_by = $%d", argIndex)
		args = append(args, changedBy)
		argIndex++
	}

	if action, ok := filters["action"].(string); ok && action != "" {
		query += fmt.Sprintf(" AND action = $%d", argIndex)
		args = append(args, action)
		argIndex++
	}

	query += " ORDER BY changed_at DESC"

	rows, err := repo.db.Master.QueryContext(ctx, query, args...)
	if err != nil {
		logger.LogRepository(func() {
			zlog.Logger.Error().
				Err(err).
				Interface("filters", filters).
				Msg("Error getting all history")
		})
		return nil, err
	}
	defer rows.Close()

	var history []models.HistoryRecord
	for rows.Next() {
		var record models.HistoryRecord
		err := rows.Scan(
			&record.ID, &record.ItemID, &record.Action,
			&record.ChangedBy, &record.ChangedAt,
		)
		if err != nil {
			logger.LogRepository(func() {
				zlog.Logger.Error().
					Err(err).
					Interface("filters", filters).
					Msg("Error scanning history record")
			})
			return nil, err
		}
		history = append(history, record)
	}

	logger.LogRepository(func() {
		zlog.Logger.Info().
			Int("records_count", len(history)).
			Interface("filters", filters).
			Msg("All history retrieved successfully")
	})
	return history, nil
}

// var _ interfaces.ItemRepository = (*ItemRepository)(nil)
