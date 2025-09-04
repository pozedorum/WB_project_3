package processor

import "time"

// ProcessingOptions настройки обработки изображения
type ProcessingOptions struct {
	Width         int
	Height        int
	Quality       int
	Format        string // "jpeg", "png", "gif"
	WatermarkText string
	Thumbnail     bool
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
