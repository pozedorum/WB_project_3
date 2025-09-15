package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/pozedorum/WB_project_3/task5/pkg/logger"
	"github.com/pozedorum/WB_project_3/task6/internal/models"
	"github.com/wb-go/wbf/dbpg"
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
		logger.LogRepository(func() { zlog.Logger.Error().Err(err).Int64("id", id).Msg("Error while processing find by id query") })
		return nil, err
	}
	logger.LogRepository(func() { zlog.Logger.Info().Int64("id", id).Msg("Sale found successfully") })
	return nil, nil
}

func (repo *SalesTrackerRepository) FindAll(ctx context.Context, filters map[string]interface{}) ([]models.SaleInformation, error) {
	return nil, nil
}
func (repo *SalesTrackerRepository) Update(ctx context.Context, id int64, sale *models.SaleInformation) error {

	return nil
}
func (repo *SalesTrackerRepository) Delete(ctx context.Context, id int64) error {
	return nil
}
func (repo *SalesTrackerRepository) GetAnalytics(ctx context.Context, req *models.AnalyticsRequest) (*models.AnalyticsResponse, error) {
	return nil, nil
}
func (repo *SalesTrackerRepository) ExportToCSV(ctx context.Context, from, to time.Time) ([]byte, error) {
	return nil, nil
}
