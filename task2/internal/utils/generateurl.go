package utils

import (
	"crypto/sha256"
	"encoding/base64"
)

const (
	codeLength = 6 // Длина короткого кода
)

func GenerateShortURL(originalURL string) string {
	hasher := sha256.New()
	hasher.Write([]byte(originalURL))
	hash := hasher.Sum(nil)

	encoded := base64.URLEncoding.EncodeToString(hash)

	shortCode := encoded[:codeLength]
	return fixShortCode(shortCode)
}

func fixShortCode(originalCode string) string {
	result := make([]byte, len(originalCode))
	for ind, char := range originalCode {
		switch {
		case char == '+' || char == '-':
			result[ind] = '0'
		case char == '/' || char == '_':
			result[ind] = 'a'
		case char >= 'a' && char <= 'z':
			result[ind] = byte(char)
		case char >= 'A' && char <= 'Z':
			result[ind] = byte(char)
		case char >= '0' && char <= '9':
			result[ind] = byte(char)
		default:
			result[ind] = 'z' // замена непонятных символы
		}
	}
	return string(result)
}

func GenerateShortURLWithSalt(originalURL, salt string) string {
	return GenerateShortURL(originalURL + salt)
}
