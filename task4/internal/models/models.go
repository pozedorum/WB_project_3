package models

import "time"

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
	ImageData   []byte
	Options     ProcessingOptions
	CallbackURL string // URL для уведомления о завершении
}

// ImageMetadata метаданные изображения
type ImageMetadata struct {
	ID             string            `json:"id"`
	OriginalName   string            `json:"original_name"`
	FileName       string            `json:"file_name"`
	ProcessedName  string            `json:"processed_name"`
	Status         string            `json:"status"` // "uploaded", "processing", "completed", "failed"
	UploadedAt     time.Time         `json:"uploaded_at"`
	ProcessedAt    time.Time         `json:"processed_at"`
	ProcessingTime time.Duration     `json:"processing_time"`
	Width          int               `json:"width"`
	Height         int               `json:"height"`
	Size           int64             `json:"size"`
	Format         string            `json:"format"`
	Options        ProcessingOptions `json:"options"`
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
