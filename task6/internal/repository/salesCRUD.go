package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/pozedorum/WB_project_3/task5/pkg/logger"
	"github.com/pozedorum/WB_project_3/task6/internal/models"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"
)

type SalesTrackerRepository struct {
	db *dbpg.DB
}

func NewSalesTrackerRepositoryWithDB(masterDSN string, slaveDSNs []string, opts *dbpg.Options) (*SalesTrackerRepository, error) {
	logger.LogRepository(func() {
		zlog.Logger.Info().Msg("Initializing SalesTrackerRepository with master and slaves")
	})
	db, err := dbpg.New(masterDSN, slaveDSNs, opts)
	if err != nil {
		logger.LogRepository(func() {
			zlog.Logger.Error().Err(err).Msg("Failed to initialize SalesTrackerRepository")
		})
		return nil, err
	}
	logger.LogRepository(func() {
		zlog.Logger.Info().Msg("SalesTrackerRepository initialized successfully")
	})
	return NewSalesTrackerRepository(db), nil
}

func NewSalesTrackerRepository(db *dbpg.DB) *SalesTrackerRepository {
	logger.LogRepository(func() {
		zlog.Logger.Info().Msg("Creating new SalesTrackerRepository instance")
	})
	return &SalesTrackerRepository{db: db}
}

func (repo *SalesTrackerRepository) Close() error {
	logger.LogRepository(func() {
		zlog.Logger.Info().Msg("Closing SalesTrackerRepository database connections")
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
		zlog.Logger.Info().Msg("SalesTrackerRepository PostgreSQL connections closed successfully")
	})
	return nil
}

func (repo *SalesTrackerRepository) Create(ctx context.Context, sale *models.SaleInformation) error {
	logger.LogRepository(func() {
		zlog.Logger.Info().
			Str("description", sale.Description).
			Str("amount", sale.Amount.String()).
			Str("type", sale.Type).
			Str("category", sale.Category).
			Time("date", sale.Date).
			Msg("Creating new sale")
	})

	insertQuery := `INSERT INTO sales (amount, type, category, description, date) 
	VALUES ($1, $2, $3, $4, $5) 
	RETURNING id, created_at, updated_at`

	err := repo.db.Master.QueryRowContext(ctx, insertQuery,
		sale.Amount, sale.Type, sale.Category,
		sale.Description, sale.Date).Scan(&sale.ID, &sale.CreatedAt, &sale.UpdatedAt)
	if err != nil {
		logger.LogRepository(func() {
			zlog.Logger.Error().
				Err(err).
				Str("description", sale.Description).
				Msg("Failed to create sale")
		})
		return err
	}

	logger.LogRepository(func() {
		zlog.Logger.Info().
			Int64("id", sale.ID).
			Str("description", sale.Description).
			Msg("Sale created successfully")
	})
	return nil
}

func (repo *SalesTrackerRepository) FindByID(ctx context.Context, id int64) (*models.SaleInformation, error) {
	logger.LogRepository(func() {
		zlog.Logger.Info().
			Int64("id", id).
			Msg("Finding sale by ID")
	})

	selectQuery := `SELECT id, amount, type, category, description, date, created_at, updated_at
	FROM sales WHERE id = $1`

	var result models.SaleInformation

	err := repo.db.Master.QueryRowContext(ctx, selectQuery, id).Scan(&result.ID, &result.Amount, &result.Type,
		&result.Category, &result.Description, &result.Date, &result.CreatedAt, &result.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.LogRepository(func() {
				zlog.Logger.Warn().
					Int64("id", id).
					Msg("Sale not found")
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
			Str("description", result.Description).
			Msg("Sale found successfully")
	})
	return &result, nil
}

func (repo *SalesTrackerRepository) FindAll(ctx context.Context, filters map[string]interface{}) ([]models.SaleInformation, error) {
	logger.LogRepository(func() {
		zlog.Logger.Info().
			Interface("filters", filters).
			Msg("Finding all sales with filters")
	})

	query := `
        SELECT id, amount, type, category, description, date, created_at, updated_at
        FROM sales WHERE 1=1
    `
	args := []interface{}{}
	argIndex := 1

	// Безопасное добавление фильтров
	if from, ok := filters["from"].(time.Time); ok {
		query += fmt.Sprintf(" AND date >= $%d", argIndex)
		args = append(args, from)
		argIndex++
	}

	if to, ok := filters["to"].(time.Time); ok {
		query += fmt.Sprintf(" AND date <= $%d", argIndex)
		args = append(args, to)
		argIndex++
	}

	if category, ok := filters["category"].(string); ok && category != "" {
		query += fmt.Sprintf(" AND category = $%d", argIndex)
		args = append(args, category)
		argIndex++
	}

	if saleType, ok := filters["type"].(string); ok && saleType != "" {
		query += fmt.Sprintf(" AND type = $%d", argIndex)
		args = append(args, saleType)
		argIndex++
	}

	// Сортировка и лимит
	query += " ORDER BY date DESC, id DESC"
	if limit, ok := filters["limit"].(int); ok && limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, limit)
	}

	var sales []models.SaleInformation

	err := retry.Do(func() error {
		logger.LogRepository(func() {
			zlog.Logger.Debug().
				Str("query", query).
				Interface("args", args).
				Msg("Executing FindAll query")
		})

		rows, err := repo.db.Master.QueryContext(ctx, query, args...)
		defer func() {
			if err := rows.Close(); err != nil {
				logger.LogRepository(func() {
					zlog.Logger.Error().
						Err(err).
						Msg("Error closing query rows")
				})
			}
		}()
		if err != nil {
			return err
		}

		for rows.Next() {
			var sale models.SaleInformation
			err := rows.Scan(
				&sale.ID,
				&sale.Amount,
				&sale.Type,
				&sale.Category,
				&sale.Description,
				&sale.Date,
				&sale.CreatedAt,
				&sale.UpdatedAt,
			)
			if err != nil {
				return err
			}
			sales = append(sales, sale)
		}
		return rows.Err()
	}, models.StandardStrategy)

	if err != nil {
		logger.LogRepository(func() {
			zlog.Logger.Error().
				Err(err).
				Interface("filters", filters).
				Msg("Error in FindAll query")
		})
		return nil, err
	}

	logger.LogRepository(func() {
		zlog.Logger.Info().
			Int("count", len(sales)).
			Interface("filters", filters).
			Msg("Sales retrieved successfully")
	})
	return sales, nil
}

func (repo *SalesTrackerRepository) Update(ctx context.Context, id int64, sale *models.SaleInformation) error {
	logger.LogRepository(func() {
		zlog.Logger.Info().
			Int64("id", id).
			Str("description", sale.Description).
			Msg("Updating sale")
	})

	updateQuery := `
        UPDATE sales 
        SET amount = $1, type = $2, category = $3, description = $4, date = $5, updated_at = NOW()
        WHERE id = $6
        RETURNING updated_at
    `

	return retry.Do(func() error {
		// Начинаем транзакцию
		tx, err := repo.db.Master.BeginTx(ctx, nil)
		if err != nil {
			logger.LogRepository(func() {
				zlog.Logger.Error().
					Err(err).
					Int64("id", id).
					Msg("Failed to begin transaction for update")
			})
			return err
		}
		defer func() {
			if err := tx.Rollback(); err != nil {
				logger.LogRepository(func() { zlog.Logger.Error().Err(err).Msg("func repo.Update failed transaction") })
			}
		}()
		// Выполняем обновление
		var updatedAt time.Time
		err = tx.QueryRowContext(ctx, updateQuery,
			sale.Amount, sale.Type, sale.Category, sale.Description, sale.Date, id,
		).Scan(&updatedAt)
		if err != nil {
			logger.LogRepository(func() {
				zlog.Logger.Error().
					Err(err).
					Int64("id", id).
					Msg("Failed to execute update query")
			})
			return err
		}

		// Коммитим транзакцию
		if err := tx.Commit(); err != nil {
			logger.LogRepository(func() {
				zlog.Logger.Error().
					Err(err).
					Int64("id", id).
					Msg("Failed to commit transaction for update")
			})
			return err
		}

		sale.UpdatedAt = updatedAt

		logger.LogRepository(func() {
			zlog.Logger.Info().
				Int64("id", id).
				Time("updated_at", updatedAt).
				Msg("Sale updated successfully")
		})
		return nil
	}, models.StandardStrategy)
}

func (repo *SalesTrackerRepository) Delete(ctx context.Context, id int64) error {
	logger.LogRepository(func() {
		zlog.Logger.Info().
			Int64("id", id).
			Msg("Deleting sale")
	})

	deleteQuery := `DELETE FROM sales WHERE id = $1`

	return retry.Do(func() error {
		tx, err := repo.db.Master.BeginTx(ctx, nil)
		if err != nil {
			logger.LogRepository(func() {
				zlog.Logger.Error().
					Err(err).
					Int64("id", id).
					Msg("Failed to begin transaction for delete")
			})
			return err
		}
		defer func() {
			if err := tx.Rollback(); err != nil {
				logger.LogRepository(func() { zlog.Logger.Error().Err(err).Msg("func repo.Delete failed transaction") })
			}
		}()

		result, err := tx.ExecContext(ctx, deleteQuery, id)
		if err != nil {
			logger.LogRepository(func() {
				zlog.Logger.Error().
					Err(err).
					Int64("id", id).
					Msg("Failed to execute delete query")
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
					Msg("No rows affected when deleting sale - sale not found")
			})
			return sql.ErrNoRows
		}

		if err := tx.Commit(); err != nil {
			logger.LogRepository(func() {
				zlog.Logger.Error().
					Err(err).
					Int64("id", id).
					Msg("Failed to commit transaction for delete")
			})
			return err
		}

		logger.LogRepository(func() {
			zlog.Logger.Info().
				Int64("id", id).
				Int64("rows_affected", rowsAffected).
				Msg("Sale deleted successfully")
		})
		return nil
	}, models.StandardStrategy)
}
