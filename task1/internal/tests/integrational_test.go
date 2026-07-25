package tests

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/pozedorum/WB_project_3/task1/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	repository "github.com/pozedorum/WB_project_3/task1/internal/repository/postgres"
)

func setupPostgres(t *testing.T) (*postgres.PostgresContainer, string) {
	t.Helper()

	ctx := context.Background()

	container, err := postgres.Run(
		ctx,
		"postgres:16",
		postgres.WithDatabase("notifications"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(time.Minute),
		),
	)

	require.NoError(t, err)

	conn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	runMigrations(t, conn)

	return container, conn
}

func runMigrations(t *testing.T, dsn string) {
	t.Helper()

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)

	defer db.Close()

	migration, err := os.ReadFile("../../migrations/001_init_notifications.sql")
	require.NoError(t, err)

	_, err = db.Exec(string(migration))
	require.NoError(t, err)
}

func TestNotificationRepository_CreateAndGet(t *testing.T) {

	container, dsn := setupPostgres(t)
	defer container.Terminate(context.Background())

	repo, err := repository.NewNotificationRepositoryWithDB(
		dsn,
		nil,
		nil,
	)

	require.NoError(t, err)

	ctx := context.Background()

	notification := &models.Notification{
		ID:        "test-id",
		UserID:    "user-1",
		Message:   "hello",
		Channel:   "email",
		Status:    models.StatusPending,
		SendAt:    time.Now().Add(time.Minute),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = repo.CreateNotification(
		ctx,
		notification,
	)

	require.NoError(t, err)

	result, err := repo.GetByID(
		ctx,
		"test-id",
	)

	require.NoError(t, err)

	require.NotNil(t, result)

	assert.Equal(
		t,
		notification.ID,
		result.ID,
	)

	assert.Equal(
		t,
		notification.Message,
		result.Message,
	)
}
