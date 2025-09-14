package server

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pozedorum/WB_project_3/task5/pkg/logger"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
)

// JWTConfig конфигурация JWT

// JWTAuthMiddleware middleware для проверки JWT токена
func (serv *EventBookerServer) JWTAuthMiddleware() ginext.HandlerFunc {
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
		userID, err := serv.parseJWTToken(tokenString)
		if err != nil {
			c.JSON(401, ginext.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Сохраняем userID в контекст для использования в хэндлерах
		c.Set("userID", userID)
		c.Next()
	}
}

// parseJWTToken парсит и валидирует JWT токен
func (serv *EventBookerServer) parseJWTToken(tokenString string) (int, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(serv.jwtConfig.SecretKey), nil
	})

	if err != nil {
		logger.LogServer(func() { zlog.Logger.Error().Err(err).Msg("Failed to parse JWT token") })
		return 0, fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		logger.LogServer(func() { zlog.Logger.Warn().Msg("Invalid JWT token") })
		return 0, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		logger.LogServer(func() { zlog.Logger.Error().Msg("Failed to extract claims from JWT token") })
		return 0, fmt.Errorf("invalid token claims")
	}

	// Извлекаем userID
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		logger.LogServer(func() { zlog.Logger.Error().Msg("User ID not found in JWT claims") })
		return 0, fmt.Errorf("user ID not found in token")
	}
	return int(userIDFloat), nil
}

// generateJWTToken создает JWT токен для пользователя (только для использования в хэндлерах)
func (serv *EventBookerServer) generateJWTToken(userID int) (string, time.Time, error) {
	// Вычисляем время истечения
	expiresAt := time.Now().Add(serv.jwtConfig.TokenLifespan)

	// Создаем claims с данными пользователя
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     expiresAt.Unix(),  // ✅ UNIX timestamp в секундах
		"iat":     time.Now().Unix(), // ✅ UNIX timestamp в секундах
	}

	// Создаем токен с claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Подписываем токен с использованием секретного ключа
	tokenString, err := token.SignedString([]byte(serv.jwtConfig.SecretKey))
	if err != nil {
		logger.LogServer(func() { zlog.Logger.Error().Err(err).Int("user_id", userID).Msg("Failed to sign JWT token") })
		return "", time.Time{}, fmt.Errorf("failed to sign token: %w", err)
	}

	logger.LogServer(func() {
		zlog.Logger.Info().
			Int("user_id", userID).
			Time("expires_at", expiresAt).
			Int64("exp_unix", expiresAt.Unix()).
			Msg("JWT token generated successfully")
	})

	return tokenString, expiresAt, nil
}
