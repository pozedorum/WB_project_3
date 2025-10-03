package interfaces

import (
	"context"

	"github.com/pozedorum/WB_project_3/task7/internal/models"
	"github.com/wb-go/wbf/ginext"
)

type Repository interface {
	Create(ctx context.Context, item *models.Item) error
	FindByID(ctx context.Context, id int64) (*models.Item, error)
	FindAll(ctx context.Context) ([]models.Item, error)
	Update(ctx context.Context, id int64, item *models.Item) error
	Delete(ctx context.Context, id int64) error
	GetItemHistory(ctx context.Context, itemID int64) ([]models.HistoryRecord, error)
	GetAllHistory(ctx context.Context, filters map[string]interface{}) ([]models.HistoryRecord, error)
}

type AuthService interface {
	GenerateToken(username string, role models.UserRole) (string, error)
	ValidateToken(token string) (*models.JWTClaims, error)
	HasPermission(role models.UserRole, action string) bool
}

type Service interface {
	CreateItem(ctx context.Context, req *models.ItemRequest, username string, role models.UserRole) (*models.Item, error)
	GetItemByID(ctx context.Context, id int64) (*models.Item, error)
	GetAllItems(ctx context.Context) ([]models.Item, error)
	UpdateItem(ctx context.Context, id int64, req *models.ItemRequest, username string, role models.UserRole) (*models.Item, error)
	DeleteItem(ctx context.Context, id int64, username string, role models.UserRole) error
	GetItemHistory(ctx context.Context, itemID int64, role models.UserRole) ([]models.HistoryRecord, error)
	GetAllHistory(ctx context.Context, filters map[string]interface{}, role models.UserRole) ([]models.HistoryRecord, error)
}

type Server interface {
	CreateItem(c *ginext.Context)
	GetItems(c *ginext.Context)
	GetItemByID(c *ginext.Context)
	UpdateItem(c *ginext.Context)
	DeleteItem(c *ginext.Context)
	GetItemHistory(c *ginext.Context)
	GetAllHistory(c *ginext.Context)
	ExportHistory(c *ginext.Context)

	Login(c *ginext.Context)
	SetupRoutes(router *ginext.Engine, apiRouter *ginext.RouterGroup)
}

type Closer interface {
	Close() error
}
