package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pozedorum/WB_project_3/task1/internal/models"
	"github.com/pozedorum/WB_project_3/task1/internal/testsutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNotificationService_Create(t *testing.T) {
	tests := []struct {
		name        string
		channel     string
		repoErr     error
		queueErr    error
		wantErr     bool
		errContains string
	}{
		{
			name:    "success",
			channel: "email",
		},
		{
			name:        "repository error",
			channel:     "email",
			repoErr:     errors.New("db error"),
			wantErr:     true,
			errContains: "failed to create notification",
		},
		{
			name:        "queue error",
			channel:     "email",
			queueErr:    errors.New("rabbitmq error"),
			wantErr:     true,
			errContains: "failed to publish to queue",
		},
		{
			name:        "unsupported channel",
			channel:     "telegram",
			wantErr:     true,
			errContains: "unsupported channel",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(testsutils.MockRepository)
			cache := new(testsutils.MockCache)
			queue := new(testsutils.MockQueue)
			notifier := &testsutils.MockNotifier{Channel: "email"}

			service := NewNotificationService(
				repo,
				cache,
				queue,
				[]Notifier{notifier},
			)

			req := &models.CreateNotificationRequest{
				UserID:  "user-1",
				Message: "hello",
				Channel: tt.channel,
				SendAt:  time.Now().Add(2 * time.Minute),
			}

			if tt.channel == "email" {
				repo.
					On("CreateNotification", mock.Anything, mock.AnythingOfType("*models.Notification")).
					Return(tt.repoErr)

				if tt.repoErr == nil {
					cache.
						On("Set", mock.Anything, mock.Anything, mock.Anything).
						Return(nil)

					queue.
						On("PublishWithDelay",
							mock.Anything,
							"notifications",
							mock.Anything,
							mock.AnythingOfType("time.Duration")).
						Return(tt.queueErr)

					if tt.queueErr != nil {
						repo.
							On("UpdateNotificationStatus",
								mock.Anything,
								mock.AnythingOfType("string"),
								models.StatusFailed).
							Return(nil)
					}
				}
			}

			notification, err := service.Create(context.Background(), req)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, notification)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
				require.NotNil(t, notification)

				assert.NotEmpty(t, notification.ID)
				assert.Equal(t, models.StatusPending, notification.Status)
			}
		})
	}
}

func TestNotificationService_GetByID(t *testing.T) {
	tests := []struct {
		name    string
		repoRes *models.Notification
		repoErr error
		wantErr bool
	}{
		{
			name: "success from repository",
			repoRes: &models.Notification{
				ID:     "id-1",
				Status: models.StatusPending,
			},
		},
		{
			name:    "repository error",
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(testsutils.MockRepository)
			cache := new(testsutils.MockCache)
			queue := new(testsutils.MockQueue)

			cache.
				On("Get", mock.Anything, "id-1").
				Return((*models.Notification)(nil), errors.New("cache miss"))

			repo.
				On("GetByID", mock.Anything, "id-1").
				Return(tt.repoRes, tt.repoErr)

			if tt.repoRes != nil {
				cache.
					On("Set", mock.Anything, "id-1", mock.Anything).
					Return(nil)
			}

			service := NewNotificationService(repo, cache, queue, nil)

			n, err := service.GetByID(context.Background(), "id-1")

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, n)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.repoRes, n)
			}
		})
	}
}

func TestNotificationService_Delete(t *testing.T) {
	tests := []struct {
		name              string
		cacheNotification *models.Notification
		repoNotification  *models.Notification
		deleteErr         error
		wantErr           string
	}{
		{
			name: "success from cache",
			cacheNotification: &models.Notification{
				ID:     "id-1",
				Status: models.StatusPending,
			},
		},
		{
			name: "success from repository",
			repoNotification: &models.Notification{
				ID:     "id-1",
				Status: models.StatusPending,
			},
		},
		{
			name:    "not found",
			wantErr: "notification not found",
		},
		{
			name: "already sent",
			cacheNotification: &models.Notification{
				ID:     "id-1",
				Status: models.StatusSent,
			},
			wantErr: "cannot delete sent notification",
		},
		{
			name: "already canceled",
			cacheNotification: &models.Notification{
				ID:     "id-1",
				Status: models.StatusCanceled,
			},
			wantErr: "notification already canceled",
		},
		{
			name: "repository delete error",
			cacheNotification: &models.Notification{
				ID:     "id-1",
				Status: models.StatusPending,
			},
			deleteErr: errors.New("db error"),
			wantErr:   "failed to delete notification",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(testsutils.MockRepository)
			cache := new(testsutils.MockCache)
			queue := new(testsutils.MockQueue)

			// Первый вызов всегда идет в кэш
			cache.
				On("Get", mock.Anything, "id-1").
				Return(tt.cacheNotification, nil)

			notification := tt.cacheNotification

			// При промахе кэша сервис обращается в БД
			if notification == nil {
				repo.
					On("GetByID", mock.Anything, "id-1").
					Return(tt.repoNotification, nil)

				notification = tt.repoNotification

				// GetByID асинхронно обновляет кэш
				if notification != nil {
					cache.
						On("Set", mock.Anything, "id-1", mock.Anything).
						Return(nil)
				}
			}

			// Если дошли до удаления
			if notification != nil &&
				notification.Status == models.StatusPending {

				repo.
					On("DeleteNotification", mock.Anything, "id-1").
					Return(tt.deleteErr)

				if tt.deleteErr == nil {
					cache.
						On("Set", mock.Anything, "id-1", nil).
						Return(nil)
				}
			}

			service := NewNotificationService(repo, cache, queue, nil)

			err := service.Delete(context.Background(), "id-1")

			if tt.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			}

			repo.AssertExpectations(t)
			cache.AssertExpectations(t)
		})
	}
}

func TestNotificationService_ProcessNotification(t *testing.T) {
	tests := []struct {
		name        string
		status      string
		notifier    bool
		wantErr     bool
		errContains string
	}{
		{
			name:     "success",
			status:   models.StatusPending,
			notifier: true,
		},
		{
			name:        "not pending",
			status:      models.StatusSent,
			wantErr:     true,
			errContains: "notification is no longer pending",
		},
		{
			name:        "notifier not found",
			status:      models.StatusPending,
			wantErr:     true,
			errContains: "no notifier for channel",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(testsutils.MockRepository)
			cache := new(testsutils.MockCache)
			queue := new(testsutils.MockQueue)

			var notifiers []Notifier
			var notifier *testsutils.MockNotifier

			if tt.notifier {
				notifier = &testsutils.MockNotifier{Channel: "email"}
				notifier.On("Send", mock.Anything).Return(nil)
				notifiers = []Notifier{notifier}
			}

			service := NewNotificationService(repo, cache, queue, notifiers)

			n := &models.Notification{
				ID:      "id-1",
				Channel: "email",
				Status:  tt.status,
			}

			cache.On("Get", mock.Anything, "id-1").Return(n, nil)

			if tt.status == models.StatusPending && tt.notifier {
				repo.
					On("UpdateNotificationStatus", mock.Anything, "id-1", models.StatusSent).
					Return(nil)

				cache.
					On("Set", mock.Anything, "id-1", n).
					Return(nil)
			}

			err := service.ProcessNotification(context.Background(), n)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNotificationService_ProcessNotificationData_InvalidJSON(t *testing.T) {
	service := NewNotificationService(
		new(testsutils.MockRepository),
		new(testsutils.MockCache),
		new(testsutils.MockQueue),
		nil,
	)

	err := service.ProcessNotificationData(context.Background(), []byte("{"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal notification")
}

func TestNotificationService_validateRequest(t *testing.T) {
	service := &notificationService{}

	tests := []struct {
		name    string
		req     *models.CreateNotificationRequest
		wantErr string
	}{
		{
			name: "valid",
			req: &models.CreateNotificationRequest{
				UserID:  "1",
				Message: "hello",
				Channel: "email",
				SendAt:  time.Now().Add(2 * time.Minute),
			},
		},
		{
			name: "empty user",
			req: &models.CreateNotificationRequest{
				Message: "hello",
				Channel: "email",
				SendAt:  time.Now().Add(2 * time.Minute),
			},
			wantErr: "user_id is required",
		},
		{
			name: "empty message",
			req: &models.CreateNotificationRequest{
				UserID:  "1",
				Channel: "email",
				SendAt:  time.Now().Add(2 * time.Minute),
			},
			wantErr: "message is required",
		},
		{
			name: "empty channel",
			req: &models.CreateNotificationRequest{
				UserID:  "1",
				Message: "hello",
				SendAt:  time.Now().Add(2 * time.Minute),
			},
			wantErr: "channel is required",
		},
		{
			name: "invalid send time",
			req: &models.CreateNotificationRequest{
				UserID:  "1",
				Message: "hello",
				Channel: "email",
				SendAt:  time.Now(),
			},
			wantErr: "send_at must be at least 1 minute in the future",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateRequest(tt.req)

			if tt.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tt.wantErr)
			}
		})
	}
}
