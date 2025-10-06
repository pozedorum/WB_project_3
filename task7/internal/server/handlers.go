package server

import (
	"strconv"

	"github.com/pozedorum/WB_project_3/task7/internal/models"
	"github.com/wb-go/wbf/ginext"
)

func (serv *WarehouseServer) CreateItem(c *ginext.Context) {
	username, role, ok := serv.GetUserFromContext(c)
	if !ok {
		c.JSON(401, ginext.H{"error": "User not found in context"})
		return
	}

	var req models.ItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, ginext.H{"error": err.Error()})
		return
	}

	item, err := serv.service.CreateItem(c.Request.Context(), &req, username, role)
	if err != nil {
		if err == models.ErrNotEnoughRights {
			c.JSON(models.StatusForbidden, ginext.H{"error": err.Error()})
			return
		}
		c.JSON(500, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(201, item)
}

func (serv *WarehouseServer) GetItems(c *ginext.Context) {
	items, err := serv.service.GetAllItems(c.Request.Context())
	if err != nil {
		c.JSON(500, ginext.H{"error": err.Error()})
		return
	}
	c.JSON(200, items)
}

func (serv *WarehouseServer) GetItemByID(c *ginext.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(400, ginext.H{"error": "Invalid ID"})
		return
	}

	item, err := serv.service.GetItemByID(c.Request.Context(), id)
	if err != nil {
		if err == models.ErrNoRows {
			c.JSON(404, ginext.H{"error": "Item not found"})
			return
		}
		c.JSON(500, ginext.H{"error": err.Error()})
		return
	}
	c.JSON(200, item)
}

func (serv *WarehouseServer) UpdateItem(c *ginext.Context) {
	username, role, ok := serv.GetUserFromContext(c)
	if !ok {
		c.JSON(401, ginext.H{"error": "User not found in context"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(400, ginext.H{"error": "Invalid ID"})
		return
	}

	var req models.ItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, ginext.H{"error": err.Error()})
		return
	}

	item, err := serv.service.UpdateItem(c.Request.Context(), id, &req, username, role)
	if err != nil {
		if err == models.ErrNoRows {
			c.JSON(404, ginext.H{"error": "Item not found"})
			return
		}
		c.JSON(500, ginext.H{"error": err.Error()})
		return
	}
	c.JSON(200, item)
}

func (serv *WarehouseServer) DeleteItem(c *ginext.Context) {
	username, role, ok := serv.GetUserFromContext(c)
	if !ok {
		c.JSON(401, ginext.H{"error": "User not found in context"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(400, ginext.H{"error": "Invalid ID"})
		return
	}

	err = serv.service.DeleteItem(c.Request.Context(), id, username, role)
	if err != nil {
		if err == models.ErrNoRows {
			c.JSON(404, ginext.H{"error": "Item not found"})
			return
		}
		c.JSON(500, ginext.H{"error": err.Error()})
		return
	}
	c.JSON(204, nil)
}

func (serv *WarehouseServer) GetItemHistory(c *ginext.Context) {
	_, role, ok := serv.GetUserFromContext(c)
	if !ok {
		c.JSON(401, ginext.H{"error": "User not found in context"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(400, ginext.H{"error": "Invalid ID"})
		return
	}

	history, err := serv.service.GetItemHistory(c.Request.Context(), id, role)
	if err != nil {
		c.JSON(500, ginext.H{"error": err.Error()})
		return
	}
	c.JSON(200, history)
}

func (serv *WarehouseServer) GetAllHistory(c *ginext.Context) {
	_, role, ok := serv.GetUserFromContext(c)
	if !ok {
		c.JSON(401, ginext.H{"error": "User not found in context"})
		return
	}

	filters := make(map[string]interface{})
	if changedBy := c.Query("changed_by"); changedBy != "" {
		filters["changed_by"] = changedBy
	}
	if action := c.Query("action"); action != "" {
		filters["action"] = action
	}

	history, err := serv.service.GetAllHistory(c.Request.Context(), filters, role)
	if err != nil {
		c.JSON(500, ginext.H{"error": err.Error()})
		return
	}
	c.JSON(200, history)
}

func (serv *WarehouseServer) ExportHistory(c *ginext.Context) {
	c.JSON(501, ginext.H{"error": "Not implemented"})
}

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
