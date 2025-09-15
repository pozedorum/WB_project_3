package repository

import (
	"context"
	"time"

	"github.com/pozedorum/WB_project_3/task5/pkg/logger"
	"github.com/pozedorum/WB_project_3/task6/internal/models"
	"github.com/shopspring/decimal"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/zlog"
)

type AnalyticsTrackerRepository struct {
	db *dbpg.DB
}

func NewAnalyticsTrackerRepositoryWithDB(masterDSN string, slaveDSNs []string, opts *dbpg.Options) (*AnalyticsTrackerRepository, error) {
	db, err := dbpg.New(masterDSN, slaveDSNs, opts)
	if err != nil {
		return nil, err
	}
	return NewAnalyticsTrackerRepository(db), nil
}

func NewAnalyticsTrackerRepository(db *dbpg.DB) *AnalyticsTrackerRepository {
	return &AnalyticsTrackerRepository{db: db}
}

func (repo *AnalyticsTrackerRepository) Close() {
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

func (repo *AnalyticsTrackerRepository) GetSalesSummary(ctx context.Context, from, to time.Time, category, saleType string) (decimal.Decimal, decimal.Decimal, int64, error) {
	return decimal.Decimal{}, decimal.Decimal{}, 0, nil
}
func (repo *AnalyticsTrackerRepository) GetMedian(ctx context.Context, from, to time.Time, category, saleType string) (decimal.Decimal, error) {
	return decimal.Decimal{}, nil
}
func (repo *AnalyticsTrackerRepository) GetPercentile90(ctx context.Context, from, to time.Time, category, saleType string) (decimal.Decimal, error) {
	return decimal.Decimal{}, nil
}
func (repo *AnalyticsTrackerRepository) GetGroupedData(ctx context.Context, from, to time.Time, groupBy, category, saleType string) ([]models.GroupedDataItem, error) {
	return nil, nil
}
