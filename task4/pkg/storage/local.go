package storage

import (
	"context"
	"fmt"
	"mime"
	"os"
	"path/filepath"
)

const (
	dirPermissionMode  = 0755
	filePermissionMode = 0644
)

type LocalStorage struct {
	basePath string
	baseURL  string
}

func NewLocalStorage(basePath, baseURL string) (*LocalStorage, error) {
	if err := os.MkdirAll(basePath, dirPermissionMode); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %v", err)
	}
	return &LocalStorage{basePath: basePath, baseURL: baseURL}, nil
}

func (ls *LocalStorage) Save(ctx context.Context, data []byte, filename string) (*FileInfo, error) {
	pathToFile := filepath.Join(ls.basePath, filename)
	dir := filepath.Dir(pathToFile)
	if err := os.MkdirAll(dir, dirPermissionMode); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %v", err)
	}

	if err := os.WriteFile(pathToFile, data, filePermissionMode); err != nil {
		return nil, fmt.Errorf("failed to save file: %v", err)
	}

	info, err := os.Stat(pathToFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %v", err)
	}
	ext := filepath.Ext(pathToFile)
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream" // бинарный файл неизвестного типа
	}

	return &FileInfo{
		Path:     pathToFile,
		Size:     info.Size(),
		MimeType: mimeType,
	}, nil
}

func (ls *LocalStorage) Get(ctx context.Context, filename string) ([]byte, error) {
	pathToFile := filepath.Join(ls.basePath, filename)
	return os.ReadFile(pathToFile)
}

func (ls *LocalStorage) Delete(ctx context.Context, filename string) error {
	pathToFile := filepath.Join(ls.basePath, filename)
	return os.Remove(pathToFile)
}
func (ls *LocalStorage) Exists(ctx context.Context, filename string) (bool, error) {
	filePath := filepath.Join(ls.basePath, filename)
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}
func (ls *LocalStorage) GetURL(ctx context.Context, filename string) (string, error) {
	return fmt.Sprintf("%s/%s", ls.baseURL, filename), nil
}
