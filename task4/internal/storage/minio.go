package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pozedorum/WB_project_3/task4/internal/models"
)

// MinIOStorage реализация хранилища для MinIO
type MinIOStorage struct {
	client     *minio.Client
	bucketName string
	region     string
}

// NewMinIOStorage создает новое MinIO хранилище
func NewMinIOStorage(Endpoint, AccessKeyID, SecretAccessKey, BucketName, Region string, UseSSL bool) (*MinIOStorage, error) {
	// Инициализация MinIO клиента
	client, err := minio.New(Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(AccessKeyID, SecretAccessKey, ""),
		Secure: UseSSL,
		Region: Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	// Проверяем соединение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	exists, err := client.BucketExists(ctx, BucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	// Создаем bucket если не существует
	if !exists {
		err = client.MakeBucket(ctx, BucketName, minio.MakeBucketOptions{
			Region: Region,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}

		// Настраиваем политику доступа (публичный доступ для чтения)
		policy := `{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Allow",
					"Principal": {"AWS": ["*"]},
					"Action": ["s3:GetObject"],
					"Resource": ["arn:aws:s3:::%s/*"]
				}
			]
		}`

		err = client.SetBucketPolicy(ctx, BucketName, fmt.Sprintf(policy, BucketName))
		if err != nil {
			return nil, fmt.Errorf("failed to set bucket policy: %w", err)
		}
	}

	return &MinIOStorage{
		client:     client,
		bucketName: BucketName,
		region:     Region,
	}, nil
}

// Save сохраняет файл в MinIO
func (m *MinIOStorage) Save(ctx context.Context, data []byte, filename string) (*models.FileInfo, error) {
	// Определяем MIME type
	ext := filepath.Ext(filename)
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// Загружаем файл
	reader := bytes.NewReader(data)
	info, err := m.client.PutObject(ctx, m.bucketName, filename, reader, int64(len(data)), minio.PutObjectOptions{
		ContentType: mimeType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload to MinIO: %w", err)
	}

	return &models.FileInfo{
		Path:     filename,
		Size:     info.Size,
		MimeType: mimeType,
		ETag:     info.ETag,
	}, nil
}

// Get загружает файл из MinIO
func (m *MinIOStorage) Get(ctx context.Context, filename string) ([]byte, error) {
	object, err := m.client.GetObject(ctx, m.bucketName, filename, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}
	defer object.Close()

	data, err := io.ReadAll(object)
	if err != nil {
		return nil, fmt.Errorf("failed to read object data: %w", err)
	}

	return data, nil
}

// Delete удаляет файл из MinIO
func (m *MinIOStorage) Delete(ctx context.Context, filename string) error {
	err := m.client.RemoveObject(ctx, m.bucketName, filename, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}
	return nil
}

// Exists проверяет существование файла
func (m *MinIOStorage) Exists(ctx context.Context, filename string) (bool, error) {
	_, err := m.client.StatObject(ctx, m.bucketName, filename, minio.StatObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("failed to check object existence: %w", err)
	}
	return true, nil
}

// GetURL возвращает URL для доступа к файлу
func (m *MinIOStorage) GetURL(ctx context.Context, filename string) (string, error) {
	// Presigned URL действителен 24 часа
	url, err := m.client.PresignedGetObject(ctx, m.bucketName, filename, 24*time.Hour, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return url.String(), nil
}

// GetFileInfo возвращает информацию о файле
func (m *MinIOStorage) GetFileInfo(ctx context.Context, filename string) (*models.FileInfo, error) {
	info, err := m.client.StatObject(ctx, m.bucketName, filename, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object info: %w", err)
	}

	// Определяем MIME type из метаданных или расширения
	mimeType := info.ContentType
	if mimeType == "" {
		ext := filepath.Ext(filename)
		mimeType = mime.TypeByExtension(ext)
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}
	}

	return &models.FileInfo{
		Path:     filename,
		Size:     info.Size,
		MimeType: mimeType,
		ETag:     info.ETag,
	}, nil
}

// ListFiles возвращает список файлов
func (m *MinIOStorage) ListFiles(ctx context.Context, prefix string) ([]models.FileInfo, error) {
	var files []models.FileInfo

	// Создаем канал для получения объектов
	objectCh := m.client.ListObjects(ctx, m.bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	// Обрабатываем объекты
	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", object.Err)
		}

		files = append(files, models.FileInfo{
			Path: object.Key,
			Size: object.Size,
			ETag: object.ETag,
		})
	}

	return files, nil
}

// CreateFolder создает виртуальную папку
func (m *MinIOStorage) CreateFolder(ctx context.Context, path string) error {
	// В S3 папки создаются путем загрузки пустого объекта с "/" в конце
	folderPath := path + "/"
	_, err := m.client.PutObject(ctx, m.bucketName, folderPath, bytes.NewReader([]byte{}), 0, minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to create folder: %w", err)
	}
	return nil
}

//var _ service.Storage = (*MinIOStorage)(nil)
