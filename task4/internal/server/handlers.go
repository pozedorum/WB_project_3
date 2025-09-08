package server

import (
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pozedorum/WB_project_3/task4/internal/models"
	"github.com/pozedorum/wbf/ginext"
	"github.com/pozedorum/wbf/zlog" // Добавляем импорт для логирования
)

func (ips *ImageProcessServer) UploadNewImage(c *ginext.Context) {

	imageHeader, err := c.FormFile("image")
	if err != nil {
		c.JSON(models.StatusBadRequest, ginext.H{
			"error":   "No image file provided",
			"details": "Please provide an image file using 'image' form field",
		})
		return
	}

	// Проверка на размер
	if imageHeader.Size > 10<<20 {
		c.JSON(models.StatusBadRequest, ginext.H{
			"error":   "File too large",
			"details": "Maximum file size is 10MB",
		})
		return
	}

	// Получение файла
	imageFile, err := imageHeader.Open()
	if err != nil {
		c.JSON(models.StatusInternalServerError, ginext.H{
			"error": "Failed to open uploaded file",
		})
		return
	}
	defer imageFile.Close()

	// Прочтение информации из файла
	imageData, err := io.ReadAll(imageFile)
	if err != nil {
		c.JSON(models.StatusInternalServerError, ginext.H{
			"error": "Failed to read file data",
		})
		return
	}

	// 4. Проверяем, что это изображение
	if !isValidImage(imageHeader.Filename, imageData) {
		c.JSON(models.StatusBadRequest, ginext.H{
			"error":   "Invalid image file",
			"details": "Please provide a valid JPEG, PNG, or GIF image",
		})
		return
	}
	zlog.Logger.Debug().
		Interface("query_params", c.Request.URL.Query()).
		Msg("Query parameters")

	form := c.Request.Form
	zlog.Logger.Debug().
		Interface("form_values", form).
		Msg("Form values")
	// 5. Парсим опции обработки
	opts, callbackURL, err := parseProcessingOptions(c)
	if err != nil {
		c.JSON(models.StatusBadRequest, ginext.H{
			"error":   "Invalid processing options",
			"details": err.Error(),
		})
		return
	}

	// Логируем полученные опции на уровне handler
	zlog.Logger.Info().
		Str("filename", imageHeader.Filename).
		Int("file_size", len(imageData)).
		Interface("processing_options", opts).
		Msg("Received upload request with processing options")

	// 6. Вызываем сервис
	result, err := ips.service.UploadImage(c.Request.Context(), imageData, imageHeader.Filename, opts, callbackURL)
	if err != nil {
		zlog.Logger.Error().
			Err(err).
			Str("filename", imageHeader.Filename).
			Msg("Failed to upload image")
		c.JSON(models.StatusInternalServerError, ginext.H{
			"error":   "Failed to upload image",
			"details": err.Error(),
		})
		return
	}

	// 7. Возвращаем результат
	zlog.Logger.Info().
		Str("image_id", result.ImageID).
		Msg("Image upload accepted for processing")
	c.JSON(models.StatusAccepted, result) // 202 Accepted - задача принята в обработку
}

func (ips *ImageProcessServer) GetProcessedImage(c *ginext.Context) {
	imageID := c.Param("id")
	if imageID == "" {
		c.JSON(models.StatusBadRequest, ginext.H{
			"error": "Image ID is required",
		})
		return
	}

	zlog.Logger.Info().
		Str("image_id", imageID).
		Msg("Request for processed image")

	result, err := ips.service.GetImage(c.Request.Context(), imageID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			zlog.Logger.Warn().
				Str("image_id", imageID).
				Msg("Image not found")
			c.JSON(models.StatusNotFound, ginext.H{
				"error":    "Image not found",
				"image_id": imageID,
			})
		} else {
			zlog.Logger.Error().
				Err(err).
				Str("image_id", imageID).
				Msg("Failed to get image")
			c.JSON(models.StatusInternalServerError, ginext.H{
				"error":   "Failed to get image",
				"details": err.Error(),
			})
		}
		return
	}

	// Если изображение еще обрабатывается
	if result.Metadata.Status != "completed" {
		zlog.Logger.Info().
			Str("image_id", imageID).
			Str("status", result.Metadata.Status).
			Msg("Image is still processing")
		c.JSON(models.StatusOK, ginext.H{
			"status":   result.Metadata.Status,
			"image_id": result.Metadata.ID,
			"message":  "Image is still being processed",
			"metadata": result.Metadata,
		})
		return
	}

	contentType := getContentType(result.Metadata.Format)
	zlog.Logger.Info().
		Str("image_id", imageID).
		Str("format", result.Metadata.Format).
		Int("size", len(result.ImageData)).
		Msg("Returning processed image data")

	c.Data(models.StatusOK, contentType, result.ImageData)
}

func (ips *ImageProcessServer) DeleteImage(c *ginext.Context) {
	imageID := c.Param("id")
	if imageID == "" {
		c.JSON(models.StatusBadRequest, ginext.H{
			"error": "Image ID is required",
		})
		return
	}

	zlog.Logger.Info().
		Str("image_id", imageID).
		Msg("Request to delete image")

	err := ips.service.DeleteImage(c.Request.Context(), imageID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			zlog.Logger.Warn().
				Str("image_id", imageID).
				Msg("Image not found for deletion")
			c.JSON(models.StatusNotFound, ginext.H{
				"error":    "Image not found",
				"image_id": imageID,
			})
		} else {
			zlog.Logger.Error().
				Err(err).
				Str("image_id", imageID).
				Msg("Failed to delete image")
			c.JSON(models.StatusInternalServerError, ginext.H{
				"error":   "Failed to delete image",
				"details": err.Error(),
			})
		}
		return
	}

	zlog.Logger.Info().
		Str("image_id", imageID).
		Msg("Image deleted successfully")

	c.JSON(models.StatusOK, ginext.H{
		"message":  "Image deleted successfully",
		"image_id": imageID,
	})
}

func (ips *ImageProcessServer) HealthCheck(c *ginext.Context) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "comment-tree-service",
	}

	zlog.Logger.Debug().Msg("Health check completed")
	c.JSON(models.StatusOK, response)
}

func (ips *ImageProcessServer) ServeFrontend(c *ginext.Context) {
	c.HTML(models.StatusOK, "index.html", nil)
}

// isValidImage проверяет, что файл является изображением
func isValidImage(filename string, data []byte) bool {
	// Проверяем по расширению
	ext := strings.ToLower(filepath.Ext(filename))
	validExtensions := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
	}
	if !validExtensions[ext] {
		return false
	}

	// Простая проверка сигнатур файлов
	if len(data) < 4 {
		return false
	}

	// JPEG: FF D8 FF
	if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return true
	}
	// PNG: 89 50 4E 47
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return true
	}
	// GIF: GIF87a or GIF89a
	if string(data[:6]) == "GIF87a" || string(data[:6]) == "GIF89a" {
		return true
	}

	return false
}

// getContentType возвращает Content-Type для формата изображения
func getContentType(format string) string {
	switch format {
	case "jpeg", "jpg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	default:
		return "application/octet-stream"
	}
}

// parseProcessingOptions парсит опции обработки из query string или form-data
func parseProcessingOptions(c *ginext.Context) (models.ProcessingOptions, string, error) {
	opts := models.ProcessingOptions{}
	var callbackURL string
	// Вспомогательная функция для получения значения из form-data или query string
	getParam := func(key string) string {
		if val := c.PostForm(key); val != "" {
			return val
		}
		return c.Query(key)
	}
	// Парсим callback URL
	if url := c.PostForm("callbackUrl"); url != "" {
		callbackURL = url
	}
	// Парсим ширину
	if widthStr := getParam("width"); widthStr != "" {
		width, err := strconv.Atoi(widthStr)
		if err != nil || width <= 0 {
			return opts, callbackURL, fmt.Errorf("invalid width: must be positive integer")
		}
		opts.Width = width
	}

	// Парсим высоту
	if heightStr := getParam("height"); heightStr != "" {
		height, err := strconv.Atoi(heightStr)
		if err != nil || height <= 0 {
			return opts, callbackURL, fmt.Errorf("invalid height: must be positive integer")
		}
		opts.Height = height
	}

	// Парсим качество
	if qualityStr := getParam("quality"); qualityStr != "" {
		quality, err := strconv.Atoi(qualityStr)
		if err != nil || quality < 1 || quality > 100 {
			return opts, callbackURL, fmt.Errorf("invalid quality: must be between 1 and 100")
		}
		opts.Quality = quality
	}

	// Парсим формат
	if format := getParam("format"); format != "" {
		format = strings.ToLower(format)
		if format != "jpeg" && format != "jpg" && format != "png" && format != "gif" {
			return opts, callbackURL, fmt.Errorf("invalid format: supported formats are jpeg, png, gif")
		}
		opts.Format = format
	}

	// Парсим водяной знак
	if watermark := getParam("watermark"); watermark != "" {
		opts.WatermarkText = watermark
	}

	// Парсим флаг миниатюры
	if thumbnail := getParam("thumbnail"); thumbnail != "" {
		thumb, err := strconv.ParseBool(thumbnail)
		if err != nil {
			return opts, callbackURL, fmt.Errorf("invalid thumbnail value: must be true or false")
		}
		opts.Thumbnail = thumb
	}

	// Логируем распарсенные опции
	zlog.Logger.Debug().
		Interface("parsed_options", opts).
		Msg("Parsed processing options")

	return opts, callbackURL, nil
}
