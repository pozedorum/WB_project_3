package testsutils

import (
	"context"
	"time"

	"github.com/pozedorum/WB_project_3/task1/internal/models"
	"github.com/stretchr/testify/mock"
)

// =============================================================
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateNotification(ctx context.Context, n *models.Notification) error {
	args := m.Called(ctx, n)
	return args.Error(0)
}

func (m *MockRepository) GetByID(ctx context.Context, id string) (*models.Notification, error) {
	args := m.Called(ctx, id)

	var notification *models.Notification
	if args.Get(0) != nil {
		notification = args.Get(0).(*models.Notification)
	}

	return notification, args.Error(1)
}

func (m *MockRepository) UpdateNotificationStatus(ctx context.Context, id, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockRepository) DeleteNotification(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

//=============================================================

type MockCache struct {
	mock.Mock
}

func (m *MockCache) Set(ctx context.Context, key string, value interface{}) error {
	args := m.Called(ctx, key, value)
	return args.Error(0)
}

func (m *MockCache) Get(ctx context.Context, key string) (*models.Notification, error) {
	args := m.Called(ctx, key)

	var notification *models.Notification
	if args.Get(0) != nil {
		notification = args.Get(0).(*models.Notification)
	}

	return notification, args.Error(1)
}

func (m *MockCache) Ping(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func (m *MockCache) Close() error {
	args := m.Called()
	return args.Error(0)
}

//=============================================================

type MockQueue struct {
	mock.Mock
}

func (m *MockQueue) PublishWithDelay(ctx context.Context, queueName string, message interface{}, delay time.Duration) error {
	args := m.Called(ctx, queueName, message, delay)
	return args.Error(0)
}

func (m *MockQueue) Consume(ctx context.Context, queueName string) (<-chan []byte, error) {
	args := m.Called(ctx, queueName)

	var ch <-chan []byte
	if args.Get(0) != nil {
		ch = args.Get(0).(<-chan []byte)
	}

	return ch, args.Error(1)
}

func (m *MockQueue) Close() error {
	args := m.Called()
	return args.Error(0)
}

//=============================================================

type MockNotifier struct {
	mock.Mock
	Channel string
}

func (m *MockNotifier) Send(notification *models.Notification) error {
	args := m.Called(notification)
	return args.Error(0)
}

func (m *MockNotifier) GetChannel() string {
	return m.Channel
}

//=============================================================

type MockService struct {
	mock.Mock
}

func (m *MockService) Create(ctx context.Context, req *models.CreateNotificationRequest) (*models.Notification, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Notification), args.Error(1)
}

func (m *MockService) GetByID(ctx context.Context, id string) (*models.Notification, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Notification), args.Error(1)
}

func (m *MockService) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockService) Consume(ctx context.Context, queue string) (<-chan []byte, error) {
	args := m.Called(ctx, queue)
	return args.Get(0).(<-chan []byte), args.Error(1)
}

func (m *MockService) ProcessNotificationData(ctx context.Context, data []byte) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *MockService) ProcessNotification(ctx context.Context, n *models.Notification) error {
	args := m.Called(ctx, n)
	return args.Error(0)
}
