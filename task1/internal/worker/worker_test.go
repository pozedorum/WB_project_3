package worker

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/pozedorum/WB_project_3/task1/internal/models"
	"github.com/pozedorum/WB_project_3/task1/internal/testsutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"testing"
	"time"
)

func TestCalculateBackoff(t *testing.T) {
	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{1, time.Second},
		{2, 2 * time.Second},
		{3, 4 * time.Second},
		{10, time.Minute},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("attempt_%d", tt.attempt), func(t *testing.T) {
			assert.Equal(t, tt.want, calculateBackoff(tt.attempt, time.Second))
		})
	}
}

func TestWorker_ProcessWithRetry(t *testing.T) {
	service := new(testsutils.MockService)

	worker := NewWorker(service)

	n := &models.Notification{
		ID: "id-1",
	}

	service.
		On("ProcessNotification", mock.Anything, n).
		Return(errors.New("fail")).
		Once()

	service.
		On("ProcessNotification", mock.Anything, n).
		Return(nil).
		Once()

	err := worker.processWithRetry(context.Background(), n)

	require.NoError(t, err)

	service.AssertExpectations(t)
}

func TestWorker_ProcessSingleMessage_InvalidBase64(t *testing.T) {
	worker := NewWorker(new(testsutils.MockService))

	worker.processSingleMessage(
		context.Background(),
		[]byte("!!!invalid base64!!!"),
	)
}

func TestWorker_ProcessSingleMessage_InvalidJSON(t *testing.T) {
	service := new(testsutils.MockService)
	worker := NewWorker(service)

	encoded := []byte(base64.StdEncoding.EncodeToString([]byte("not json")))

	worker.processSingleMessage(context.Background(), encoded)

	service.AssertNotCalled(t, "ProcessNotification")
}

func TestWorker_ProcessSingleMessage_Success(t *testing.T) {
	service := new(testsutils.MockService)
	worker := NewWorker(service)

	notification := models.Notification{
		ID:      "id-1",
		UserID:  "user-1",
		Channel: "email",
		Status:  models.StatusPending,
	}

	data, err := json.Marshal(notification)
	require.NoError(t, err)

	encoded := []byte(base64.StdEncoding.EncodeToString(data))

	service.
		On("ProcessNotification", mock.Anything, mock.AnythingOfType("*models.Notification")).
		Return(nil)

	worker.processSingleMessage(context.Background(), encoded)

	service.AssertExpectations(t)
}
