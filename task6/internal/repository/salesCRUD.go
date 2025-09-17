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
	db, err := dbpg.New(masterDSN, slaveDSNs, opts)
	if err != nil {
		return nil, err
	}
	return NewSalesTrackerRepository(db), nil
}

func NewSalesTrackerRepository(db *dbpg.DB) *SalesTrackerRepository {
	return &SalesTrackerRepository{db: db}
}

func (repo *SalesTrackerRepository) Close() {
	if err := repo.db.Master.Close(); err != nil {
		logger.LogRepository(func() { zlog.Logger.Panic().Msg("Database failed to close") })
	}
	for _, slave := range repo.db.Slaves {
		if slave != nil {
			if err := slave.Close(); err != nil {
				logger.LogRepository(func() { zlog.Logger.Panic().Msg("Slave database failed to close") })
			}
		}
	}
	logger.LogRepository(func() { zlog.Logger.Info().Msg("PostgreSQL connections closed") })
}

func (repo *SalesTrackerRepository) Create(ctx context.Context, sale *models.SaleInformation) error {
	insertQuery := `INSERT INTO sales (amount, type, category, description, date) 
	VALUES ($1, $2, $3, $4, $5) 
	RETURNING id, created_at, updated_at`

	err := repo.db.Master.QueryRowContext(ctx, insertQuery,
		sale.Amount, sale.Type, sale.Category,
		sale.Description, sale.Date).Scan(&sale.ID, &sale.CreatedAt, &sale.UpdatedAt)
	if err != nil {
		logger.LogRepository(func() { zlog.Logger.Error().Err(err).Str("description", sale.Description).Msg("Failed to create sale") })
		return err
	}
	logger.LogRepository(func() { zlog.Logger.Info().Str("description", sale.Description).Msg("Sale created successfully") })
	return nil
}

func (repo *SalesTrackerRepository) FindByID(ctx context.Context, id int64) (*models.SaleInformation, error) {
	selectQuery := `SELECT id, amount, type, category, description, date, created_at, updated_at
	FROM sales WHERE id = $1`

	var result models.SaleInformation

	err := repo.db.Master.QueryRowContext(ctx, selectQuery, id).Scan(&result.ID, &result.Amount, &result.Type,
		&result.Category, &result.Description, &result.Date, &result.CreatedAt, &result.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.LogRepository(func() { zlog.Logger.Warn().Int64("id", id).Msg("Sale not found") })
			return nil, sql.ErrNoRows
		}
		logger.LogRepository(func() { zlog.Logger.Error().Err(err).Int64("id", id).Msg("Error in FindByID query") })
		return nil, err
	}
	logger.LogRepository(func() { zlog.Logger.Info().Int64("id", id).Msg("Sale found successfully") })
	return &result, nil
}

func (repo *SalesTrackerRepository) FindAll(ctx context.Context, filters map[string]interface{}) ([]models.SaleInformation, error) {
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
		argIndex++
	}

	var sales []models.SaleInformation

	err := retry.Do(func() error {
		rows, err := repo.db.Master.QueryContext(ctx, query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

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
	}, models.StandartStrategy)

	if err != nil {
		logger.LogRepository(func() {
			zlog.Logger.Error().Err(err).Msg("Error in FindAll query")
		})
		return nil, err
	}

	logger.LogRepository(func() {
		zlog.Logger.Info().Int("count", len(sales)).Msg("Sales retrieved successfully")
	})
	return sales, nil
}

func (repo *SalesTrackerRepository) Update(ctx context.Context, id int64, sale *models.SaleInformation) error {
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
			return err
		}
		defer tx.Rollback()

		// Выполняем обновление
		var updatedAt time.Time
		err = tx.QueryRowContext(ctx, updateQuery,
			sale.Amount, sale.Type, sale.Category, sale.Description, sale.Date, id,
		).Scan(&updatedAt)
		if err != nil {
			return err
		}

		// Коммитим транзакцию
		if err := tx.Commit(); err != nil {
			return err
		}

		sale.UpdatedAt = updatedAt
		return nil
	}, models.StandartStrategy)
}
func (repo *SalesTrackerRepository) Delete(ctx context.Context, id int64) error {
	deleteQuery := `DELETE FROM sales WHERE id = $1`

	return retry.Do(func() error {
		tx, err := repo.db.Master.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		result, err := tx.ExecContext(ctx, deleteQuery, id)
		if err != nil {
			return err
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}

		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		return tx.Commit()
	}, models.StandartStrategy)
}
