package server

import (
	"github.com/pozedorum/WB_project_3/task7/internal/models"
	"github.com/wb-go/wbf/ginext"
)

func (serv *WarehouseServer) CreateItem(c *ginext.Context)
func (serv *WarehouseServer) GetItems(c *ginext.Context)
func (serv *WarehouseServer) GetItemByID(c *ginext.Context)
func (serv *WarehouseServer) UpdateItem(c *ginext.Context)
func (serv *WarehouseServer) DeleteItem(c *ginext.Context)
func (serv *WarehouseServer) GetItemHistory(c *ginext.Context)
func (serv *WarehouseServer) GetAllHistory(c *ginext.Context)
func (serv *WarehouseServer) ExportHistory(c *ginext.Context)

// Login хэндлер для входа (выбор роли)
func (serv *WarehouseServer) Login(c *ginext.Context) {
	type LoginRequest struct {
		Username string          `json:"username" binding:"required"`
		Role     models.UserRole `json:"role" binding:"required,oneof=admin manager viewer"`
	}

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, ginext.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Генерируем токен
	token, expiresAt, err := serv.jwtConfig.GenerateJWTToken(req.Username, req.Role)
	if err != nil {
		c.JSON(500, ginext.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(200, ginext.H{
		"token":      token,
		"expires_at": expiresAt,
		"username":   req.Username,
		"role":       req.Role,
	})
}

// GetProfile хэндлер для получения профиля пользователя
func (serv *WarehouseServer) GetProfile(c *ginext.Context) {
	username, role, ok := serv.GetUserFromContext(c)
	if !ok {
		c.JSON(401, ginext.H{"error": "User not found in context"})
		return
	}

	c.JSON(200, ginext.H{
		"username": username,
		"role":     role,
	})
}
