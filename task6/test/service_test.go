package test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pozedorum/WB_project_3/task6/internal/models"
	"github.com/pozedorum/WB_project_3/task6/internal/service"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// Mock репозиториев
type MockSaleRepository struct {
	mock.Mock
}

func (m *MockSaleRepository) Create(ctx context.Context, sale *models.SaleInformation) error {
	args := m.Called(ctx, sale)
	return args.Error(0)
}

func (m *MockSaleRepository) FindByID(ctx context.Context, id int64) (*models.SaleInformation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SaleInformation), args.Error(1)
}

func (m *MockSaleRepository) FindAll(ctx context.Context, filters map[string]interface{}) ([]models.SaleInformation, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]models.SaleInformation), args.Error(1)
}

func (m *MockSaleRepository) Update(ctx context.Context, id int64, sale *models.SaleInformation) error {
	args := m.Called(ctx, id, sale)
	return args.Error(0)
}

func (m *MockSaleRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSaleRepository) Close() error {
	args := m.Called()
	return args.Error(0)
}

type MockAnalyticsRepository struct {
	mock.Mock
}

func (m *MockAnalyticsRepository) GetAnalytics(ctx context.Context, req *models.AnalyticsRequest) (*models.AnalyticsResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AnalyticsResponse), args.Error(1)
}

func (m *MockAnalyticsRepository) GetSalesSummary(ctx context.Context, req *models.AnalyticsRequest) (*models.SalesSummaryResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SalesSummaryResponse), args.Error(1)
}

func (m *MockAnalyticsRepository) GetMedian(ctx context.Context, req *models.AnalyticsRequest) (decimal.Decimal, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

func (m *MockAnalyticsRepository) GetPercentile90(ctx context.Context, req *models.AnalyticsRequest) (decimal.Decimal, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

func (m *MockAnalyticsRepository) ExportToCSV(ctx context.Context, req *models.AnalyticsRequest) ([]byte, error) {
	args := m.Called(ctx, req)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockAnalyticsRepository) Close() error {
	args := m.Called()
	return args.Error(0)
}

type ServiceTestSuite struct {
	suite.Suite
	saleRepo      *MockSaleRepository
	analyticsRepo *MockAnalyticsRepository
	service       *service.SaleTrackerService
}

func TestServiceSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}

func (suite *ServiceTestSuite) SetupTest() {
	suite.saleRepo = new(MockSaleRepository)
	suite.analyticsRepo = new(MockAnalyticsRepository)
	suite.service = service.New(suite.saleRepo, suite.analyticsRepo)
}

func (suite *ServiceTestSuite) TestCreateSale_Success() {
	t := suite.T()

	req := &models.SaleRequest{
		Amount:      decimal.NewFromFloat(100.50),
		Type:        "income",
		Category:    "test-category",
		Description: "Test description",
		Date:        time.Now().AddDate(0, 0, -1),
	}

	suite.saleRepo.On("Create", mock.Anything, mock.MatchedBy(func(sale *models.SaleInformation) bool {
		// Проверяем, что все поля правильно установлены (кроме ID и временных меток)
		return sale.Amount.Equal(req.Amount) &&
			sale.Type == req.Type &&
			sale.Category == req.Category &&
			sale.Description == req.Description &&
			sale.Date.Equal(req.Date) &&
			!sale.CreatedAt.IsZero() &&
			!sale.UpdatedAt.IsZero()
	})).Return(nil)

	result, err := suite.service.CreateSale(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Проверяем основные поля
	assert.True(t, result.Amount.Equal(req.Amount))
	assert.Equal(t, req.Type, result.Type)
	assert.Equal(t, req.Category, result.Category)
	assert.Equal(t, req.Description, result.Description)
	assert.True(t, result.Date.Equal(req.Date))

	// Временные метки должны быть установлены
	assert.False(t, result.CreatedAt.IsZero())
	assert.False(t, result.UpdatedAt.IsZero())

	suite.saleRepo.AssertExpectations(t)
}

func (suite *ServiceTestSuite) TestCreateSale_ValidationError() {
	testCases := []struct {
		name string
		req  *models.SaleRequest
	}{
		{
			name: "Negative amount",
			req: &models.SaleRequest{
				Amount:      decimal.NewFromFloat(-100),
				Type:        "income",
				Category:    "test",
				Description: "Test",
				Date:        time.Now(),
			},
		},
		{
			name: "Invalid type",
			req: &models.SaleRequest{
				Amount:      decimal.NewFromFloat(100),
				Type:        "invalid",
				Category:    "test",
				Description: "Test",
				Date:        time.Now(),
			},
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			result, err := suite.service.CreateSale(context.Background(), tc.req)

			assert.Error(t, err)
			assert.Nil(t, result)
		})
	}
}

func (suite *ServiceTestSuite) TestGetSaleByID_NotFound() {
	suite.saleRepo.On("FindByID", mock.Anything, int64(999)).Return(nil, models.ErrSaleNotFound)

	result, err := suite.service.GetSaleByID(context.Background(), 999)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.True(suite.T(), errors.Is(err, models.ErrSaleNotFound))
}
