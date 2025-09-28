package test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	di "github.com/pozedorum/WB_project_3/task6/internal/DI"
	"github.com/pozedorum/WB_project_3/task6/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/wb-go/wbf/zlog"
)

type IntegrationTestSuite struct {
	suite.Suite
	container *di.Container
	server    *httptest.Server
	client    *http.Client
}

func TestIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.Run(t, new(IntegrationTestSuite))
}

func (suite *IntegrationTestSuite) SetupSuite() {
	// Инициализация логгера
	zlog.Init()

	// Загрузка конфигурации
	cfg := config.Load()

	// Создание контейнера
	container, err := di.NewContainer(cfg)
	assert.NoError(suite.T(), err)

	suite.container = container

	// Создание тестового сервера
	router := suite.setupRouter()
	suite.server = httptest.NewServer(router)

	suite.client = &http.Client{
		Timeout: 30 * time.Second,
	}

	fmt.Println("Test server started on:", suite.server.URL)
}

func (suite *IntegrationTestSuite) TearDownSuite() {
	if suite.server != nil {
		suite.server.Close()
	}

	if suite.container != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		suite.container.Shutdown(ctx)
	}
}

func (suite *IntegrationTestSuite) setupRouter() http.Handler {
	// Используем роутер из контейнера
	router := suite.container.HTTPServer.Handler
	return router
}
