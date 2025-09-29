Отлично! Вот комплексные запросы для тестирования вашего API, соответствующие всем моделям и тегам.

## **🚀 Запросы для тестирования CRUD операций**

### **1. Создание записей (POST /items)**

#### **Доход (income):**
```bash
curl -X POST http://localhost:8080/items \
  -H "Content-Type: application/json" \
  -d '{
    "amount": "15000.50",
    "type": "income",
    "category": "salary",
    "description": "Monthly salary",
    "date": "2024-01-15T10:00:00Z"
  }'
```

#### **Расход (expense):**
```bash
curl -X POST http://localhost:8080/items \
  -H "Content-Type: application/json" \
  -d '{
    "amount": "2500.75",
    "type": "expense", 
    "category": "food",
    "description": "Groceries for week",
    "date": "2024-01-16T18:30:00Z"
  }'
```

#### **Еще несколько тестовых записей:**
```bash
# Доход - фриланс
curl -X POST http://localhost:8080/items \
  -H "Content-Type: application/json" \
  -d '{
    "amount": "8000.00",
    "type": "income",
    "category": "freelance",
    "description": "Website development",
    "date": "2024-01-10T14:20:00Z"
  }'

# Расход - транспорт
curl -X POST http://localhost:8080/items \
  -H "Content-Type: application/json" \
  -d '{
    "amount": "1500.00",
    "type": "expense",
    "category": "transport",
    "description": "Monthly transport pass",
    "date": "2024-01-05T08:15:00Z"
  }'

# Расход - развлечения
curl -X POST http://localhost:8080/items \
  -H "Content-Type: application/json" \
  -d '{
    "amount": "3000.00", 
    "type": "expense",
    "category": "entertainment",
    "description": "Cinema and dinner",
    "date": "2024-01-12T20:45:00Z"
  }'
```

### **2. Получение всех записей (GET /items)**
```bash
curl http://localhost:8080/items
```

### **3. Получение записи по ID (GET /items/{id})**
```bash
# Замените {id} на реальный ID из ответа предыдущего запроса
curl http://localhost:8080/items/1
curl http://localhost:8080/items/2
```

### **4. Обновление записи (PUT /items/{id})**
```bash
curl -X PUT http://localhost:8080/items/1 \
  -H "Content-Type: application/json" \
  -d '{
    "amount": "16000.00",
    "type": "income",
    "category": "salary",
    "description": "Salary with bonus",
    "date": "2024-01-15T10:00:00Z"
  }'
```

### **5. Удаление записи (DELETE /items/{id})**
```bash
curl -X DELETE http://localhost:8080/items/3
```

## **📊 Запросы для тестирования аналитики (GET /analytics)**

### **1. Базовая аналитика за период:**
```bash
curl "http://localhost:8080/analytics?from=2024-01-01T00:00:00Z&to=2024-01-31T23:59:59Z"
```

### **2. Аналитика только по доходам:**
```bash
curl "http://localhost:8080/analytics?from=2024-01-01T00:00:00Z&to=2024-01-31T23:59:59Z&type=income"
```

### **3. Аналитика только по расходам:**
```bash
curl "http://localhost:8080/analytics?from=2024-01-01T00:00:00Z&to=2024-01-31T23:59:59Z&type=expense"
```

### **4. Аналитика по конкретной категории:**
```bash
curl "http://localhost:8080/analytics?from=2024-01-01T00:00:00Z&to=2024-01-31T23:59:59Z&category=food"
```

### **5. Аналитика с группировкой по дням:**
```bash
curl "http://localhost:8080/analytics?from=2024-01-01T00:00:00Z&to=2024-01-31T23:59:59Z&group_by=day"
```

### **6. Аналитика с группировкой по неделям:**
```bash
curl "http://localhost:8080/analytics?from=2024-01-01T00:00:00Z&to=2024-01-31T23:59:59Z&group_by=week"
```

### **7. Аналитика с группировкой по месяцам:**
```bash
curl "http://localhost:8080/analytics?from=2024-01-01T00:00:00Z&to=2024-01-31T23:59:59Z&group_by=month"
```

### **8. Аналитика с группировкой по категориям:**
```bash
curl "http://localhost:8080/analytics?from=2024-01-01T00:00:00Z&to=2024-01-31T23:59:59Z&group_by=category"
```

### **9. Комбинированная аналитика:**
```bash
curl "http://localhost:8080/analytics?from=2024-01-01T00:00:00Z&to=2024-01-31T23:59:59Z&type=expense&group_by=category"
```

## **📁 Экспорт в CSV (GET /csv)**
```bash
# Скачивание CSV файла
curl http://localhost:8080/csv -o sales_export.csv

# Просмотр в терминале
curl http://localhost:8080/csv
```

## **🎯 Тестирование валидации и ошибок**

### **Неправильный тип операции:**
```bash
curl -X POST http://localhost:8080/items \
  -H "Content-Type: application/json" \
  -d '{
    "amount": "1000.00",
    "type": "invalid_type",
    "category": "test",
    "date": "2024-01-15T10:00:00Z"
  }'
```

### **Отсутствует обязательное поле:**
```bash
curl -X POST http://localhost:8080/items \
  -H "Content-Type: application/json" \
  -d '{
    "type": "income",
    "category": "test",
    "date": "2024-01-15T10:00:00Z"
  }'
```

### **Неправильная группировка в аналитике:**
```bash
curl "http://localhost:8080/analytics?from=2024-01-01T00:00:00Z&to=2024-01-31T23:59:59Z&group_by=invalid_group"
```

### **Отсутствуют даты в аналитике:**
```bash
curl "http://localhost:8080/analytics"
```
