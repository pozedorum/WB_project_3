package server

import (
	"strings"

	"github.com/pozedorum/WB_project_3/task5/pkg/logger"
	"github.com/pozedorum/WB_project_3/task7/internal/models"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
)

// JWTAuthMiddleware middleware для проверки JWT токена
func (serv *WarehouseServer) JWTAuthMiddleware() ginext.HandlerFunc {
	return func(c *ginext.Context) {
		// Извлекаем токен из заголовка Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, ginext.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Проверяем формат "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(401, ginext.H{"error": "Authorization header format must be 'Bearer <token>'"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Парсим и валидируем токен
		username, userRole, err := serv.jwtConfig.ParseJWTToken(tokenString)
		if err != nil {
			c.JSON(401, ginext.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Сохраняем данные пользователя в контекст
		c.Set("username", username)
		c.Set("role", userRole)
		c.Next()
	}
}

// RoleMiddleware middleware для проверки ролей
func (serv *WarehouseServer) RoleMiddleware(allowedRoles ...models.UserRole) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		// Получаем роль из контекста
		roleRaw, exists := c.Get("role")
		if !exists {
			c.JSON(401, ginext.H{"error": "Role not found in context"})
			c.Abort()
			return
		}

		role, ok := roleRaw.(models.UserRole)
		if !ok {
			c.JSON(401, ginext.H{"error": "Invalid role type"})
			c.Abort()
			return
		}

		// Проверяем, есть ли роль в списке разрешенных
		hasAccess := false
		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				hasAccess = true
				break
			}
		}

		if !hasAccess {
			logger.LogServer(func() {
				zlog.Logger.Warn().
					Str("role", string(role)).
					Interface("allowed_roles", allowedRoles).
					Msg("Access denied: insufficient permissions")
			})
			c.JSON(403, ginext.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetUserFromContext вспомогательная функция для получения данных пользователя из контекста
func (serv *WarehouseServer) GetUserFromContext(c *ginext.Context) (username string, role models.UserRole, ok bool) {
	usernameRaw, exists := c.Get("username")
	if !exists {
		return "", "", false
	}

	roleRaw, exists := c.Get("role")
	if !exists {
		return "", "", false
	}

	username, ok1 := usernameRaw.(string)
	role, ok2 := roleRaw.(models.UserRole)

	return username, role, ok1 && ok2
}
