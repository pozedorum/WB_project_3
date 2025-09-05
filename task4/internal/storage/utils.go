package storage

import (
	"crypto/rand"
	"fmt"
	"mime"
	"path/filepath"
	"strings"
	"time"
)

// GenerateFilename генерирует уникальное имя файла
func GenerateFilename(originalName string) string {
	ext := filepath.Ext(originalName)
	timestamp := time.Now().UnixNano()

	// Добавляем случайность
	randomBytes := make([]byte, 4)
	rand.Read(randomBytes)
	randomPart := fmt.Sprintf("%x", randomBytes)

	return fmt.Sprintf("%d_%s%s", timestamp, randomPart, ext)
}

// GetContentType определяет MIME type по расширению файла
func GetContentType(filename string) string {
	ext := filepath.Ext(filename)
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		return "application/octet-stream"
	}
	return mimeType
}

// IsImage проверяет, является ли файл изображением
func IsImage(filename string) bool {
	contentType := GetContentType(filename)
	return strings.HasPrefix(contentType, "image/")
}

// ValidFilename проверяет валидность имени файла
func ValidFilename(filename string) bool {
	// Запрещенные символы в S3
	forbiddenChars := []string{"\\", "..", "~", "^", "{", "}", "[", "]", "%", "#", "|", "<", ">", "\x00"}

	for _, char := range forbiddenChars {
		if strings.Contains(filename, char) {
			return false
		}
	}

	// Максимальная длина 1024 символа
	if len(filename) > 1024 {
		return false
	}

	return true
}

// GetFileExtension возвращает расширение файла из MIME type
func GetFileExtension(mimeType string) string {
	extensions, err := mime.ExtensionsByType(mimeType)
	if err != nil || len(extensions) == 0 {
		return ".bin"
	}
	return extensions[0] // возвращаем первое подходящее расширение
}

// NormalizePath нормализует путь для S3
func NormalizePath(path string) string {
	// Убираем ведущие и завершающие слеши
	path = strings.Trim(path, "/")
	// Заменяем множественные слеши на одинарные
	return strings.ReplaceAll(path, "//", "/")
}
