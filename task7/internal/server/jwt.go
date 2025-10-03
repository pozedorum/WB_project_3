package server

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pozedorum/WB_project_3/task7/internal/models"
)

// JWTConfig конфигурация JWT
type JWTConfig struct {
	SecretKey     string
	TokenLifespan time.Duration
}

// parseJWTToken парсит и валидирует JWT токен
func (conf *JWTConfig) ParseJWTToken(tokenString string) (string, models.UserRole, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(conf.SecretKey), nil
	})

	if err != nil || !token.Valid {
		return "", "", fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", fmt.Errorf("invalid token claims")
	}

	username, ok1 := claims["username"].(string)
	roleStr, ok2 := claims["role"].(string)

	if !ok1 || !ok2 {
		return "", "", fmt.Errorf("invalid token data")
	}

	return username, models.UserRole(roleStr), nil
}

// generateJWTToken создает JWT токен для пользователя
func (conf *JWTConfig) GenerateJWTToken(username string, role models.UserRole) (string, time.Time, error) {
	expiresAt := time.Now().Add(conf.TokenLifespan)

	claims := jwt.MapClaims{
		"username": username,
		"role":     string(role),
		"exp":      expiresAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(conf.SecretKey))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}
