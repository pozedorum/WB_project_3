package models

import (
	"time"

	"github.com/wb-go/wbf/retry"
)

type FileInfo struct {
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type"`
	ETag     string `json:"etag"`
}

// ProcessingResult результат обработки
type ProcessingResult struct {
	ProcessedData  []byte
	Format         string
	Width          int // Длина
	Height         int // Ширина
	Size           int64
	ProcessingTime time.Duration
}

// ProcessingTask задача на обработку
type ProcessingTask struct {
	TaskID      string
	Options     ProcessingOptions
	CallbackURL string // URL для уведомления о завершении
}

// CallbackNotification структура для уведомления
type CallbackNotification struct {
	TaskID    string `json:"task_id"`
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Message   string `json:"message,omitempty"`
}

// ImageMetadata метаданные изображения
type ImageMetadata struct {
	ID            string            `json:"id"`
	OriginalName  string            `json:"original_name"`
	FileName      string            `json:"file_name"`
	ProcessedName string            `json:"processed_name"`
	Status        string            `json:"status"` // "uploaded", "processing", "completed", "failed"
	UploadedAt    time.Time         `json:"uploaded_at"`
	ProcessedAt   time.Time         `json:"processed_at"`
	Width         int               `json:"width"`
	Height        int               `json:"height"`
	Size          int64             `json:"size"`
	Format        string            `json:"format"`
	Options       ProcessingOptions `json:"options"`
}

func (im *ImageMetadata) GetProcessingTime() time.Duration {
	if im.UploadedAt.IsZero() || im.ProcessedAt.IsZero() {
		return 0
	}
	return im.ProcessedAt.Sub(im.UploadedAt)
}

// ProcessingMessage структура сообщения для Kafka
type ProcessingMessage struct {
	TaskID      string            `json:"task_id"`
	Options     ProcessingOptions `json:"options"`
	CallbackURL string            `json:"callback_url,omitempty"`
	Timestamp   time.Time         `json:"timestamp"`
}

// UploadResult результат загрузки
type UploadResult struct {
	ImageID string `json:"image_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// ImageResult результат получения изображения
type ImageResult struct {
	ImageData []byte         `json:"image_data,omitempty"`
	ImageURL  string         `json:"image_url,omitempty"`
	Metadata  *ImageMetadata `json:"metadata"`
}

// ProcessingOptions упрощенная версия опций
type ProcessingOptions struct {
	Width         int    `json:"width"`
	Height        int    `json:"height"`
	Quality       int    `json:"quality"`
	Format        string `json:"format"`
	WatermarkText string `json:"watermark_text"`
	Thumbnail     bool   `json:"thumbnail"`
}

var (
	StandardStrategy         = retry.Strategy{Attempts: 3, Delay: time.Second}
	ProduserConsumerStrategy = retry.Strategy{Attempts: 5, Delay: 2 * time.Second}
)

const (
	StatusOK                  = 200
	StatusAccepted            = 202
	StatusFound               = 302
	StatusBadRequest          = 400
	StatusNotFound            = 404
	StatisConflict            = 409
	StatusInternalServerError = 500
)
