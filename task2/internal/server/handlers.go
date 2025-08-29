package server

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pozedorum/WB_project_3/task2/internal/models"
	"github.com/pozedorum/wbf/ginext"
	"github.com/pozedorum/wbf/zlog"
)

func (ss *ShortURLServer) Shorten(c *ginext.Context) {
	var request struct {
		URL        string `json:"url" binding:"required,url"`
		CustomCode string `json:"custom_code,omitempty" binding:"omitempty,alphanum,min=1,max=6"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		zlog.Logger.Error().Err(err).Msg("Failed to bind JSON for create short URL")
		c.JSON(models.StatusBadRequest, ginext.H{"error": "Invalid request: " + err.Error()})
		return
	}

	zlog.Logger.Info().
		Str("original_url", request.URL).
		Str("custom_code", request.CustomCode).
		Time("created_at", time.Now()).
		Msg("creating short URL")

	// Создаем короткую ссылку с учетом кастомного кода
	su, err := ss.service.CreateShortURL(c.Request.Context(), request.URL, request.CustomCode)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateShortCode) {
			zlog.Logger.Warn().
				Str("custom_code", request.CustomCode).
				Msg("Custom code already exists")
			c.JSON(models.StatisConflict, ginext.H{
				"error": "Custom short code already exists. Please choose a different one.",
			})
			return
		}

		zlog.Logger.Error().Err(err).
			Str("original_url", request.URL).
			Str("custom_code", request.CustomCode).
			Msg("Failed to create short URL")
		c.JSON(models.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	response := struct {
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}{
		ShortURL:    "/s/" + su.ShortCode,
		OriginalURL: su.OriginalURL,
	}

	zlog.Logger.Info().
		Str("original_url", response.OriginalURL).
		Str("short_url", response.ShortURL).
		Msg("Short URL created successfully")

	c.JSON(models.StatusAccepted, response)
}

func (ss *ShortURLServer) Redirect(c *ginext.Context) {
	shortCode := c.Param("shortCode")
	if shortCode == "" {
		zlog.Logger.Error().Msg("short URL not found")
		c.JSON(models.StatusBadRequest, ginext.H{"error": "short URL not found"})
		return
	}
	zlog.Logger.Info().Str("short_code", shortCode).Msg("Redirect called")
	userAgent := c.Request.UserAgent()
	ip := c.ClientIP()

	originalURL, err := ss.service.Redirect(c.Request.Context(),
		shortCode,
		userAgent,
		ip)
	if err != nil {
		if errors.Is(err, models.ErrShortURLNotFound) {
			zlog.Logger.Error().Msg("short URL not found")
			c.JSON(models.StatusBadRequest, ginext.H{"error": "short URL not found"})
			return
		}
		zlog.Logger.Error().Err(err).Str("short_code", shortCode).Msg("Failed to redirect")
		c.JSON(models.StatusBadRequest, ginext.H{"error": "Internal server error"})
		return
	}
	zlog.Logger.Info().Str("short_code", shortCode).Str("original_url", originalURL)
	c.Redirect(models.StatusFound, originalURL)
}

func (ss *ShortURLServer) Analytics(c *ginext.Context) {
	shortCode := c.Param("shortCode")

	// Опциональные параметры фильтрации
	period := c.Query("period")   // "1d", "7d", "30d"
	groupBy := c.Query("groupBy") // "day", "month", "user-agent", "browser", "os", "device"

	// Валидация параметров
	validPeriods := map[string]bool{"": true, "1d": true, "7d": true, "30d": true}
	validGroupBys := map[string]bool{"": true, "day": true, "month": true, "user-agent": true, "browser": true, "os": true, "device": true}

	if !validPeriods[period] {
		c.JSON(models.StatusBadRequest, ginext.H{"error": "Invalid period parameter. Use: 1d, 7d, 30d"})
		return
	}

	if !validGroupBys[groupBy] {
		c.JSON(models.StatusBadRequest, ginext.H{"error": "Invalid groupBy parameter. Use: day, month, user-agent, browser, os, device"})
		return
	}

	analytics, err := ss.service.GetStatByShortCode(c.Request.Context(), shortCode, period, groupBy)
	if err != nil {
		if errors.Is(err, models.ErrShortURLNotFound) {
			c.JSON(models.StatusNotFound, ginext.H{"error": "short URL not found"})
			return
		}
		zlog.Logger.Error().Err(err).Str("short_code", shortCode).Msg("Failed to get analytics")
		c.JSON(models.StatusInternalServerError, ginext.H{"error": "Failed to get analytics"})
		return
	}
	zlog.Logger.Info().Any("month_date", analytics.MonthlyStats[0])
	c.JSON(models.StatusOK, analytics)
}

func (ss *ShortURLServer) IndexPage(c *ginext.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "URL Shortener",
	})
}

func (ss *ShortURLServer) ResultPage(c *ginext.Context) {
	shortURL := c.Query("short_url")
	originalURL := c.Query("original_url")

	c.HTML(http.StatusOK, "result.html", gin.H{
		"title":        "URL Created",
		"short_url":    shortURL,
		"original_url": originalURL,
	})
}

func (ss *ShortURLServer) AnalyticsPage(c *ginext.Context) {
	c.HTML(http.StatusOK, "analytics.html", gin.H{
		"title": "Analytics",
	})
}

func (ss *ShortURLServer) HealthCheck(c *ginext.Context) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "url-shortener",
	}
	c.JSON(models.StatusOK, response)
}
