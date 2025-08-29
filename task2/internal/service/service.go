package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/pozedorum/WB_project_3/task2/internal/models"

	"github.com/pozedorum/WB_project_3/task2/internal/utils"
	"github.com/pozedorum/wbf/zlog"
)

const (
	attemptsCount = 5
)

type ShortURLService struct {
	repo Repository
}

func New(repo Repository) *ShortURLService {
	zlog.Logger.Info().Msg("Creating short url service")
	return &ShortURLService{repo: repo}
}

func (s *ShortURLService) CreateShortURL(ctx context.Context, originalURL string, customCode string) (*models.ShortURL, error) {
	if err := validateURL(originalURL); err != nil {
		return nil, err
	}

	var shortCode string

	if customCode != "" {
		shortCode = customCode
		existingURL, err := s.repo.GetOriginalURLIfExists(ctx, shortCode)
		if err == nil {
			if existingURL.OriginalURL != originalURL {
				// Кастомный код уже существует и связан с другим URL
				return nil, models.ErrDuplicateShortCode
			} else {
				// Кастомный код существует и связан с правильным URL
				return existingURL, nil
			}
		}
		if !errors.Is(err, models.ErrShortURLNotFound) {
			return nil, err
		}

	} else {
		shortCode = utils.GenerateShortURL(originalURL)
		uniqueShortCode, ok, err := s.ensureUniqueShortCode(ctx, originalURL, shortCode)
		if err != nil {
			return nil, err
		}
		if ok {
			shortURL, err := s.repo.GetOriginalURLIfExists(ctx, shortCode)
			if err != nil {
				return nil, err
			}
			return shortURL, nil
		}
		shortCode = uniqueShortCode
	}

	shortURL := &models.ShortURL{
		ShortCode:   shortCode,
		OriginalURL: originalURL,
		CreatedAt:   time.Now(),
		ClicksCount: 0,
	}

	if err := s.repo.CreateShortURL(ctx, shortURL); err != nil {
		if errors.Is(err, models.ErrDuplicateShortCode) {
			return nil, models.ErrDuplicateShortCode
		}
		return nil, err
	}

	zlog.Logger.Info().
		Str("short_code", shortCode).
		Msg("Short URL created")

	return shortURL, nil
}

func (s *ShortURLService) Redirect(ctx context.Context, shortCode string, userAgent, ip string) (string, error) {
	// Получаем полную информацию атомарно
	shortURL, err := s.repo.GetOriginalURLIfExists(ctx, shortCode)
	if err != nil {
		if errors.Is(err, models.ErrShortURLNotFound) {
			return "", models.ErrShortURLNotFound
		}
		return "", err
	}

	// Запись аналитики (асинхронно или в горутине чтобы не блокировать редирект)
	go func() {
		clickStruct := models.ClickAnalyticsEntry{
			ShortCode: shortCode,
			UserAgent: userAgent,
			IPAddress: ip,
			CreatedAt: time.Now(),
		}
		if err := s.repo.RegisterClick(context.Background(), &clickStruct); err != nil {
			zlog.Logger.Error().Err(err).Str("short_code", shortCode).Msg("Failed to register click")
		}
	}()
	zlog.Logger.Info().Str("short_code", shortCode).Str("original_url", shortURL.OriginalURL).Msg("serive layer")
	return shortURL.OriginalURL, nil
}

func (s *ShortURLService) GetStatByShortCode(ctx context.Context, shortCode string, period string, groupBy string) (*models.AnalyticsResponse, error) {
	// Проверяем существование ссылки
	_, err := s.repo.GetOriginalURLIfExists(ctx, shortCode)
	if err != nil {
		if errors.Is(err, models.ErrShortURLNotFound) {
			return nil, models.ErrShortURLNotFound
		}
		return nil, err
	}

	return s.repo.GetStatisticsByShortCode(ctx, shortCode, period, groupBy)
}

func (s *ShortURLService) ensureUniqueShortCode(ctx context.Context, originalURL, baseShortCode string) (string, bool, error) {
	// 1. Атомарно проверяем существование и получаем данные
	existingShortURL, err := s.repo.GetOriginalURLIfExists(ctx, baseShortCode)
	if err != nil {
		if errors.Is(err, models.ErrShortURLNotFound) {
			// Код свободен - возвращаем как есть
			return baseShortCode, false, nil
		}
		return "", false, fmt.Errorf("failed to check short code existence: %w", err)
	}

	// 2. Если URL совпадает - возвращаем существующий код
	if existingShortURL.OriginalURL == originalURL {
		zlog.Logger.Info().Str("short_code", baseShortCode).Msg("Returning existing short code for same URL")
		return baseShortCode, true, nil
	}

	// 3. Разные URL с одинаковым хэшем - коллизия!
	zlog.Logger.Warn().
		Str("base_short_code", baseShortCode).
		Str("existing_url", existingShortURL.OriginalURL).
		Str("new_url", originalURL).
		Msg("Hash collision detected, generating new code")

	// 4. Генерируем новые коды с солью
	for attempt := 1; attempt <= attemptsCount; attempt++ {
		saltedShortCode := utils.GenerateShortURLWithSalt(originalURL, fmt.Sprintf("%d", attempt))

		// Атомарно проверяем новый код
		existingSaltedURL, err := s.repo.GetOriginalURLIfExists(ctx, saltedShortCode)
		if err != nil {
			if errors.Is(err, models.ErrShortURLNotFound) {
				// Код свободен - используем
				zlog.Logger.Info().
					Str("new_short_code", saltedShortCode).
					Int("attempt", attempt).
					Msg("Generated unique short code after collision")
				return saltedShortCode, false, nil
			}
			return "", false, fmt.Errorf("failed to check salted short code: %w", err)
		}

		// Если нашли существующий URL - проверяем совпадение
		if existingSaltedURL.OriginalURL == originalURL {
			zlog.Logger.Info().
				Str("short_code", saltedShortCode).
				Msg("Found existing short code for the same URL")
			return saltedShortCode, false, nil
		}
	}

	return "", false, fmt.Errorf("failed to generate unique short code after %d attempts for URL: %s", attemptsCount, originalURL)
}

func validateURL(rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("URL scheme must be http or https")
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("URL must contain host")
	}

	return nil
}
