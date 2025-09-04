package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

type GoogleDriveAutoStorage struct {
	apiKey   string
	folderID string
	baseURL  string
}

func NewGoogleDriveAutoStorage(apiKey, folderID string) *GoogleDriveAutoStorage {
	return &GoogleDriveAutoStorage{
		apiKey:   apiKey,
		folderID: folderID,
		baseURL:  "https://www.googleapis.com/drive/v3",
	}
}

// Save автоматически загружает файл в Google Drive
func (g *GoogleDriveAutoStorage) Save(ctx context.Context, data []byte, filename string) (*FileInfo, error) {
	// Создаем метаданные файла
	metadata := map[string]interface{}{
		"name":    filename,
		"parents": []string{g.folderID},
	}

	metadataJSON, _ := json.Marshal(metadata)

	// Создаем multipart запрос
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Добавляем метаданные
	metadataPart, _ := writer.CreateFormField("metadata")
	metadataPart.Write(metadataJSON)

	// Добавляем файл
	filePart, _ := writer.CreateFormFile("file", filename)
	filePart.Write(data)

	writer.Close()

	// URL для загрузки
	uploadURL := fmt.Sprintf("%s/files?uploadType=multipart&key=%s", g.baseURL, g.apiKey)

	req, _ := http.NewRequest("POST", uploadURL, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("upload failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("upload failed with status: %d", resp.StatusCode)
	}

	// Парсим ответ
	var result struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	return &FileInfo{
		Path: result.ID,
		Size: int64(len(data)),
	}, nil
}

// Get автоматически скачивает файл
func (g *GoogleDriveAutoStorage) Get(ctx context.Context, fileID string) ([]byte, error) {
	downloadURL := fmt.Sprintf("%s/files/%s?alt=media&key=%s", g.baseURL, fileID, g.apiKey)

	resp, err := http.Get(downloadURL)
	if err != nil {
		return nil, fmt.Errorf("download failed: %v", err)
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// Delete автоматически удаляет файл
func (g *GoogleDriveAutoStorage) Delete(ctx context.Context, fileID string) error {
	deleteURL := fmt.Sprintf("%s/files/%s?key=%s", g.baseURL, fileID, g.apiKey)

	req, _ := http.NewRequest("DELETE", deleteURL, nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("delete failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("delete failed with status: %d", resp.StatusCode)
	}

	return nil
}

// Exists проверяет существование файла
func (g *GoogleDriveAutoStorage) Exists(ctx context.Context, fileID string) (bool, error) {
	infoURL := fmt.Sprintf("%s/files/%s?fields=id&key=%s", g.baseURL, fileID, g.apiKey)

	resp, err := http.Get(infoURL)
	if err != nil {
		return false, fmt.Errorf("check failed: %v", err)
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

// GetURL возвращает URL для просмотра
func (g *GoogleDriveAutoStorage) GetURL(ctx context.Context, fileID string) (string, error) {
	return fmt.Sprintf("https://drive.google.com/file/d/%s/view", fileID), nil
}
