package notifier

import (
	"fmt"
	"net/smtp"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pozedorum/WB_project_3/task1/internal/config"
	"github.com/pozedorum/WB_project_3/task1/internal/models"
	"github.com/pozedorum/wbf/zlog"
)

const (
	EmailType    = "email"
	TelegramType = "telegram"
)

type EmailNotifier struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
}

type TelegramNotifier struct {
	BotToken string
}

func NewEmailNotifier(config config.EmailConfig) (*EmailNotifier, error) {
	if config.SMTPHost == "" || config.SMTPUser == "" {
		return nil, fmt.Errorf("email notifier configuration incomplete")
	}
	zlog.Logger.Info().Msgf("Email notifier initialized for %s:%d", config.SMTPHost, config.SMTPPort)
	return &EmailNotifier{
		SMTPHost:     config.SMTPHost,
		SMTPPort:     config.SMTPPort,
		SMTPUser:     config.SMTPUser,
		SMTPPassword: config.SMTPPassword}, nil
}

// NewTelegramNotifier создает новый Telegram нотификатор с конфигурацией
func NewTelegramNotifier(config config.TelegramConfig) (*TelegramNotifier, error) {
	if config.BotToken == "" {
		return nil, fmt.Errorf("telegram notifier configuration incomplete")
	}
	if config.BotToken == "" {
		zlog.Logger.Info().Msg("Telegram notifier initialized without bot token (will use stub)")
	} else {
		zlog.Logger.Info().Msg("Telegram notifier initialized with bot token")
	}
	return &TelegramNotifier{BotToken: config.BotToken}, nil
}

func (en *EmailNotifier) Send(notification *models.Notification) error {
	if en.SMTPHost == "" || en.SMTPUser == "" {
		zlog.Logger.Info().Msgf("Email not configured properly, using stub for user %s", notification.UserID)
		return nil
	}

	// Реальная логика отправки для Яндекс
	auth := smtp.PlainAuth("", en.SMTPUser, en.SMTPPassword, en.SMTPHost)
	to := []string{notification.UserID}

	// Используем SMTP_USER как отправителя (Яндекс требует совпадение)
	from := en.SMTPUser

	msg := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: Notification\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n\r\n%s",
		from,
		notification.UserID,
		notification.Message,
	))

	err := smtp.SendMail(
		fmt.Sprintf("%s:%d", en.SMTPHost, en.SMTPPort),
		auth,
		from,
		to,
		msg,
	)

	if err != nil {
		zlog.Logger.Error().Err(err).Msg("Failed to send email via Yandex")
		return fmt.Errorf("failed to send email: %w", err)
	}

	zlog.Logger.Info().Msgf("✅ Email sent successfully via Yandex SMTP to %s", notification.UserID)
	return nil
}

// GetChannel возвращает тип канала для Email
func (en *EmailNotifier) GetChannel() string {
	return EmailType
}

// Send реализация для Telegram
func (tn *TelegramNotifier) Send(notification *models.Notification) error {
	zlog.Logger.Info().Msgf("📱 Attempting to send TELEGRAM to user %s: %s", notification.UserID, notification.Message)

	if tn.BotToken == "" {
		zlog.Logger.Info().Msgf("Telegram bot token not configured, using stub for user %s", notification.UserID)
		zlog.Logger.Info().Msgf("Telegram message stub sent to %s", notification.UserID)
		return nil
	}

	// Создаем бота
	bot, err := tgbotapi.NewBotAPI(tn.BotToken)
	if err != nil {
		return fmt.Errorf("failed to create telegram bot: %w", err)
	}

	// Устанавливаем режим отладки (опционально)
	bot.Debug = true

	// Конвертируем user_id в int64 (chat ID)
	chatID, err := strconv.ParseInt(notification.UserID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid telegram chat ID: %w", err)
	}

	// Создаем сообщение
	msg := tgbotapi.NewMessage(chatID, notification.Message)

	// Отправляем сообщение
	_, err = bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send telegram message: %w", err)
	}

	zlog.Logger.Info().Msgf("Telegram message sent successfully to %s", notification.UserID)
	return nil
}

// GetChannel возвращает тип канала для Telegram
func (tn *TelegramNotifier) GetChannel() string {
	return TelegramType
}
