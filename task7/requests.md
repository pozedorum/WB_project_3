# Warehouse Control System - API Documentation

## 🚀 Быстрый старт

### 1. Запуск проекта
```bash
docker-compose up --build
```


## 📋 Порядок тестирования API

### Шаг 1: Получение JWT токенов для разных ролей

#### 🔐 Получить токен для администратора
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin_user",
    "role": "admin"
  }'
```

#### 🔐 Получить токен для менеджера
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "manager_user", 
    "role": "manager"
  }'
```

#### 🔐 Получить токен для просмотрщика
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "viewer_user",
    "role": "viewer"
  }'
```

**Сохраните полученные токены для следующих запросов!**

---

### Шаг 2: Работа с товарами (требует авторизации)

#### 📦 Создать товар (только admin)
```bash
curl -X POST http://localhost:8080/items \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -d '{
    "name": "Ноутбук Dell XPS 15",
    "price": 150000
  }'
```

#### 📦 Создать еще товары
```bash
curl -X POST http://localhost:8080/items \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -d '{
    "name": "Мышь беспроводная",
    "price": 2500
  }'

curl -X POST http://localhost:8080/items \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -d '{
    "name": "Монитор 27 дюймов",
    "price": 30000
  }'
```

#### 📋 Получить все товары (все роли)
```bash
curl -X GET http://localhost:8080/items \
  -H "Authorization: Bearer YOUR_TOKEN"
```

#### 🔍 Получить конкретный товар (все роли)
```bash
curl -X GET http://localhost:8080/items/1 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

#### ✏️ Обновить товар (admin и manager)
```bash
curl -X PUT http://localhost:8080/items/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ADMIN_OR_MANAGER_TOKEN" \
  -d '{
    "name": "Ноутбук Dell XPS 15 (2024)",
    "price": 160000
  }'
```

#### 🗑️ Удалить товар (только admin)
```bash
curl -X DELETE http://localhost:8080/items/2 \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

---

### Шаг 3: Работа с историей изменений

#### 📖 Получить историю конкретного товара (все роли)
```bash
curl -X GET http://localhost:8080/history/item/1 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

#### 📚 Получить всю историю (admin и manager)
```bash
curl -X GET http://localhost:8080/history \
  -H "Authorization: Bearer YOUR_ADMIN_OR_MANAGER_TOKEN"
```

#### 🔍 Получить историю с фильтрами (admin и manager)
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

---

### Шаг 4: Профиль пользователя

#### 👤 Получить профиль текущего пользователя
```bash
curl -X GET http://localhost:8080/profile \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

## 🔐 Тестирование прав доступа

### Проверка что viewer не может создавать товары:
```bash
curl -X POST http://localhost:8080/items \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_VIEWER_TOKEN" \
  -d '{
    "name": "Тестовый товар",
    "price": 1000
  }'
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

---

## 🎯 Пример успешных ответов

### Успешный login:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_at": "2024-01-01T12:00:00Z",
  "username": "admin_user",
  "role": "admin"
}
```

### Успешное создание товара:
```json
{
  "id": 1,
  "name": "Ноутбук Dell XPS 15",
  "price": 150000,
  "created_by": "admin_user",
  "created_at": "2024-01-01T10:00:00Z",
  "updated_at": "2024-01-01T10:00:00Z"
}
```

### Успешное получение истории:
```json
[
  {
    "id": 1,
    "item_id": 1,
    "action": "CREATE",
    "changed_by": "admin_user",
    "changed_at": "2024-01-01T10:00:00Z"
  },
  {
    "id": 2,
    "item_id": 1,
    "action": "UPDATE",
    "changed_by": "admin_user",
    "changed_at": "2024-01-01T11:00:00Z"
  }
]
```

---

## 🛠️ Утилиты для тестирования

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

## ❗ Возможные проблемы и решения

1. **"Authorization header required"** - Не передан JWT токен
2. **"Invalid or expired token"** - Токен невалидный или просроченный  
3. **"Insufficient permissions"** - Недостаточно прав для действия
4. **"Item not found"** - Товар с указанным ID не существует
5. **"Invalid ID"** - Неправильный формат ID

Тестируйте запросы по порядку и убедитесь, что используете правильные токены для соответствующих ролей!