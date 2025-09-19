# Notification Service

Асинхронный сервис для отправки отложенных уведомлений через Telegram и Email.

## Features

- Отправка уведомлений через Telegram и Email
- Отложенная отправка с точным временем
- Сохранение состояния в PostgreSQL
- Асинхронная обработка через RabbitMQ
- Повторные попытки отправки при ошибках
- Веб-интерфейс для управления уведомлениями

## Prerequisites

- Docker и Docker Compose
- Go 1.19+ (для локальной разработки)
- Telegram Bot Token (для Telegram уведомлений)
- Яндекс почта с паролем приложения (для Email уведомлений)

## Quick Start

### 1. Клонирование и запуск

```bash
git clone <repository-url>
cd wbf
docker-compose up -d
```

Сервис будет доступен по адресу: http://localhost:8080

### 2. Настройка Telegram

1. Создайте бота через @BotFather
2. Получите токен бота
3. Добавьте токен в переменные окружения (добавлять в docker-compose.yml):

```bash
TELEGRAM_BOT_TOKEN=your_telegram_bot_token_here
```

### 3. Настройка Email (Яндекс)

1. Включите двухфакторную аутентификацию в Яндекс.Почте
2. Сгенерируйте пароль приложения
3. Настройте переменные окружения:

```bash
SMTP_HOST=smtp.yandex.ru
SMTP_PORT=587
SMTP_USER=your_yandex_email@yandex.ru
SMTP_PASSWORD=your_app_password_here
```

Уведомления через почту реализованы, но почему-то отказываются работать, чтобы я не делал
Тратить ещё один день на починку я посчитал слишком долгим

## Configuration

### Environment Variables

Создайте `.env` файл и настройте в `docker-compose.yml`:

```env
# Server
SERVER_PORT=8080

# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=delayed_notifier
DB_SSLMODE=disable

# Redis
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=redis
REDIS_DB=0

# RabbitMQ
RABBITMQ_HOST=rabbitmq
RABBITMQ_PORT=5672
RABBITMQ_USER=guest
RABBITMQ_PASSWORD=guest

# Email (Yandex)
SMTP_HOST=smtp.yandex.ru
SMTP_PORT=587
SMTP_USER=your_email@yandex.ru
SMTP_PASSWORD=your_app_password

```

### Docker Compose

Основные сервисы в `docker-compose.yml`:
- app: основное приложение (Go)
- postgres: база данных PostgreSQL
- redis: кэширование
- rabbitmq: очередь сообщений

## API Endpoints

### Создание уведомления
```bash
POST /notify
Content-Type: application/json

{
    "user_id": "1105031510",
    "message": "Тестовое уведомление",
    "channel": "telegram",
    "send_at": "2025-12-22T20:21:00Z"
}
```

### Получение статуса уведомления
```bash
GET /notify/{id}
```

### Удаление уведомления
```bash
DELETE /notify/{id}
```

### Health check
```bash
GET /health
```

### Проверка статуса
```bash
curl -X GET http://localhost:8080/notify/{id}
```

### Удаление уведомления
```bash
curl -X DELETE http://localhost:8080/notify/{id}
```


### Logs Monitoring

```bash
# Все логи
docker-compose logs

# Логи приложения
docker-compose logs app

# Логи в реальном времени
docker-compose logs -f app

# Поиск ошибок
docker-compose logs app | grep -i error
```



## Support

При возникновении проблем:
1. Проверьте логи контейнеров
2. Убедитесь в правильности настроек
3. Проверьте подключение к внешним сервисам (Telegram, Яндекс.Почта)

Для Telegram: убедитесь, что бот имеет права на отправку сообщений и добавлен в соответствующий чат.

Для Яндекс.Почты: убедитесь, что включена двухфакторная аутентификация и создан пароль приложения для SMTP.
