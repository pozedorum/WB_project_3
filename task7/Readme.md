# Warehouse Control System

Система управления складом с ролевой моделью доступа, историей изменений и веб-интерфейсом. Поддерживает полный цикл работы с товарами: создание, просмотр, редактирование, удаление и отслеживание истории изменений.

## Особенности

- **CRUD-операции** для управления товарами
- **Ролевая модель доступа** (admin, manager, viewer)
- **Полная история изменений** с использованием триггеров PostgreSQL
- **JWT-аутентификация** с проверкой прав доступа
- **Веб-интерфейс** для удобной работы с системой
- **Фильтрация истории** по пользователям и действиям
- **Автоматическое логирование** всех операций

## Архитектура

- **Backend**: Go с использованием Gin framework
- **Frontend**: Vanilla JavaScript + HTML/CSS
- **База данных**: PostgreSQL с триггерами для истории
- **Аутентификация**: JWT токены
- **Контейнеризация**: Docker + Docker Compose

## Быстрый старт

### 1. Клонирование и настройка

```bash
git clone <repository-url>
cd warehousecontrol
```

### 2. Настройка окружения

Создайте файл `.env` на основе примера:

```bash
cp .env.example .env
```

Отредактируйте `.env` при необходимости:

```env
# Server
SERVER_PORT=8080

# JWT
JWT_SECRET= #для генерации используйте openssl rand -hex 64
JWT_TOKEN_LIFESPAN=24h

# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=warehousecontrol
DB_SSLMODE=disable

# Retry 
MAX_RETRIES=3
BASE_DELAY=1s
```

### 3. Запуск приложения

```bash
# Сборка и запуск
make build
make run

# Или напрямую через docker-compose
docker compose up --build
```

Приложение будет доступно по адресу: http://localhost:8080

### 2. Структура базы данных

Система автоматически создает необходимые таблицы:
- `items` - таблица товаров
- `item_history` - история изменений товаров
- Триггеры для автоматического логирования операций CREATE, UPDATE, DELETE

## Ролевая модель

| Роль | Просмотр товаров | Создание товаров | Редактирование товаров | Удаление товаров | Просмотр всей истории |
|------|------------------|------------------|-----------------------|------------------|----------------------|
| **Admin** | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Manager** | ✅ | ❌ | ✅ | ❌ | ✅ |
| **Viewer** | ✅ | ❌ | ❌ | ❌ | ✅s |

## API Endpoints

### 🔐 Аутентификация

#### Получить токен для администратора
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin_user",
    "role": "admin"
  }'
```

#### Получить токен для менеджера
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "manager_user", 
    "role": "manager"
  }'
```

#### Получить токен для просмотрщика
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "viewer_user",
    "role": "viewer"
  }'
```

**Сохраните полученные токены для следующих запросов!**

### 📦 Управление товарами

#### Создать товар (только admin)
```bash
curl -X POST http://localhost:8080/items \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -d '{
    "name": "Ноутбук Dell XPS 15",
    "price": 150000
  }'
```

#### Получить все товары (все роли)
```bash
curl -X GET http://localhost:8080/items \
  -H "Authorization: Bearer YOUR_TOKEN"
```

#### Получить конкретный товар (все роли)
```bash
curl -X GET http://localhost:8080/items/1 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

#### Обновить товар (admin и manager)
```bash
curl -X PUT http://localhost:8080/items/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ADMIN_OR_MANAGER_TOKEN" \
  -d '{
    "name": "Ноутбук Dell XPS 15 (2024)",
    "price": 160000
  }'
```

#### Удалить товар (только admin)
```bash
curl -X DELETE http://localhost:8080/items/2 \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

### 📊 История изменений

#### Получить историю конкретного товара (все роли)
```bash
curl -X GET http://localhost:8080/history/item/1 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

#### Получить всю историю (admin и manager)
```bash
curl -X GET http://localhost:8080/history \
  -H "Authorization: Bearer YOUR_ADMIN_OR_MANAGER_TOKEN"
```

#### Получить историю с фильтрами
```bash
# Фильтр по пользователю
curl -X GET "http://localhost:8080/history?changed_by=admin_user" \
  -H "Authorization: Bearer YOUR_ADMIN_OR_MANAGER_TOKEN"

# Фильтр по действию
curl -X GET "http://localhost:8080/history?action=CREATE" \
  -H "Authorization: Bearer YOUR_ADMIN_OR_MANAGER_TOKEN"

# Комбинированный фильтр
curl -X GET "http://localhost:8080/history?changed_by=admin_user&action=UPDATE" \
  -H "Authorization: Bearer YOUR_ADMIN_OR_MANAGER_TOKEN"
```

### 👤 Профиль пользователя

#### Получить профиль текущего пользователя
```bash
curl -X GET http://localhost:8080/profile \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Веб-интерфейс

После запуска сервиса откройте http://localhost:8080 в браузере.

### Возможности веб-интерфейса:

- **Авторизация** с выбором роли
- **Просмотр товаров** в виде карточек
- **Создание товаров** (для admin)
- **Редактирование товаров** (для admin и manager)
- **Удаление товаров** (только для admin)
- **Просмотр истории изменений** по каждому товару
- **Фильтрация истории** по пользователям и действиям
- **Адаптивный дизайн** для разных устройств

## Примеры использования

### Полный цикл работы с товаром

1. **Войдите как администратор**
2. **Создайте несколько товаров:**
   ```bash
   curl -X POST http://localhost:8080/items \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer ADMIN_TOKEN" \
     -d '{"name": "Мышь беспроводная", "price": 2500}'
   ```

3. **Обновите товар как менеджер:**
   ```bash
   curl -X PUT http://localhost:8080/items/1 \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer MANAGER_TOKEN" \
     -d '{"name": "Мышь беспроводная Logitech", "price": 3000}'
   ```

4. **Просмотрите историю изменений:**
   ```bash
   curl -X GET http://localhost:8080/history/item/1 \
     -H "Authorization: Bearer ANY_TOKEN"
   ```

5. **Удалите товар как администратор:**
   ```bash
   curl -X DELETE http://localhost:8080/items/1 \
     -H "Authorization: Bearer ADMIN_TOKEN"
   ```

## Тестирование прав доступа

### Проверка что viewer не может создавать товары:
```bash
curl -X POST http://localhost:8080/items \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_VIEWER_TOKEN" \
  -d '{"name": "Тестовый товар", "price": 1000}'
```
**Ожидаемый ответ:** `403 Forbidden`

### Проверка что manager не может удалять товары:
```bash
curl -X DELETE http://localhost:8080/items/1 \
  -H "Authorization: Bearer YOUR_MANAGER_TOKEN"
```
**Ожидаемый ответ:** `403 Forbidden`

### Проверка что viewer не может смотреть всю историю:
```bash
curl -X GET http://localhost:8080/history \
  -H "Authorization: Bearer YOUR_VIEWER_TOKEN"
```
**Ожидаемый ответ:** `403 Forbidden`

## 🛠️ Утилиты для разработки

### Генерация JWT секрета:
```bash
openssl rand -hex 32
```

### Проверка здоровья БД:
```bash
docker-compose exec postgres pg_isready -U postgres
```

### Просмотр логов:
```bash
docker-compose logs app
docker-compose logs postgres
```

## Структура проекта

```
├── internal/
│   ├── models/          # Модели данных
│   ├── service/         # Бизнес-логика
│   ├── server/          # HTTP-обработчики и маршрутизация
│   └── frontend/        # Веб-интерфейс
├── migrations/          # Миграции базы данных
├── pkg/
│   ├── logger/          # Логирование
│   └── config/          # Конфигурация
└── docker-compose.yml   # Docker конфигурация
```

## Особенности реализации

- **Триггеры PostgreSQL** для автоматического логирования всех изменений
- **Каскадное удаление** истории при удалении товаров
- **Валидация данных** на всех уровнях приложения
- **Подробное логирование** операций с товарами
- **CORS middleware** для веб-интерфейса
- **Обработка ошибок** с понятными сообщениями

## ❗ Возможные проблемы и решения

1. **"Authorization header required"** - Не передан JWT токен
2. **"Invalid or expired token"** - Токен невалидный или просроченный  
3. **"Insufficient permissions"** - Недостаточно прав для действия
4. **"Item not found"** - Товар с указанным ID не существует
5. **"Invalid ID"** - Неправильный формат ID

## Безопасность

- JWT-аутентификация с временем жизни токенов
- Проверка прав доступа на уровне middleware
- Валидация всех входных данных
- Защита от SQL-инъекций через подготовленные запросы
- CORS политика для веб-интерфейса

---

**Примечание**: Все изменения товаров автоматически записываются в историю через триггеры базы данных. История сохраняется даже после удаления товаров.

**Примечание**: Для тестирования используйте разные токены для разных ролей, чтобы убедиться в правильной работе системы разграничения прав доступа.