# Сервис обработки изображений

Асинхронный сервис для фоновой обработки изображений с использованием Apache Kafka.

## Features

- Загрузка изображений для обработки
- Фоновая обработка через Kafka (ресайз, водяные знаки, миниатюры)
- Получение обработанных изображений
- Удаление изображений
- Веб-интерфейс для управления изображениями
- Поддержка различных опций обработки

## Prerequisites

- Docker и Docker Compose
- go 1.24.4

## Quick Start

### 1. Клонирование и запуск

```bash
git clone <repository-url>
cd wbf
make rebuild
```

Сервис будет доступен по адресу: http://localhost:8080
Файловое хранилище будет доступно по адресу: http://localhost:9001

### 2. Настройка

Создайте `.env` файл на основе `.env.example`:

```env
# Server
SERVER_PORT=8080

# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=images_db
DB_SSLMODE=disable

# Kafka
KAFKA_BROKERS=kafka:9092
KAFKA_TOPIC=image-processing-tasks
KAFKA_GROUP_ID=image-processor-group

# Storage
STORAGE_TYPE=local
STORAGE_PATH=./storage
```

### 3. Docker Compose

Основные сервисы в `docker-compose.yml`:
- image-processor: основное приложение (Go)
- minio: файловое хранилище
- kafka: Apache Kafka для обработки задач
- zookeeper: Zookeeper для Kafka

## API Endpoints

### Загрузка изображения
```bash
POST /upload
Content-Type: multipart/form-data

Параметры формы:
- image: файл изображения (обязательно)
- width: ширина (опционально)
- height: высота (опционально)
- quality: качество 1-100 (опционально)
- format: формат (jpeg, png, gif) (опционально)
- watermark: текст водяного знака (опционально)
- thumbnail: true/false для миниатюры (опционально)
```

### Получение изображения
```bash
GET /image/{id}
```

### Удаление изображения
```bash
DELETE /image/{id}
```

### Health check
```bash
GET /health
```

## Примеры использования

### Загрузка изображения с опциями обработки
```bash
# Ресайз + водяной знак
curl -X POST http://localhost:8080/upload -F "image=@./test/cat3.jpeg" -F "width=1000" -F "height=300" -F "watermark=MyCat"

# Смена формата и качества
curl -X POST http://localhost:8080/upload -F "image=@./test/cat3.jpeg" -F "format=png" -F "quality=90"

# Создание миниатюры
curl -X POST http://localhost:8080/upload -F "image=@./test/cat2.jpeg" -F "thumbnail=true"
```

### Получение обработанного изображения
Тут надо менять id на тот, что выдаёт ответ после запроса с загрузкой
```bash
# Получить изображение по ID
curl -X GET http://localhost:8080/image/1757432276029868846_741fb976.jpeg \
  --output ./test/result.jpeg

```

### Удаление изображения
```bash
# Удалить изображение по ID
curl -X DELETE http://localhost:8080/image/1757432276029868846_741fb976.jpeg
```

## Веб-интерфейс

После запуска сервиса откройте http://localhost:8080 в браузере для доступа к веб-интерфейсу.

Возможности веб-интерфейса:
- Загрузка изображений через выбор файла
- Настройка опций обработки (ресайз, формат, качество, водяные знаки)
- Просмотр статуса обработки в реальном времени
- Скачивание готовых изображений
- Удаление изображений

## Опции обработки

### Ресайз
- `width`: ширина в пикселях
- `height`: высота в пикселях
- Соотношение сторон сохраняется автоматически

### Формат и качество
- `format`: jpeg, png, gif (по умолчанию - оригинальный формат)
- `quality`: 1-100

### Водяные знаки
- `watermark`: текст водяного знака
- Знак добавляется в правый нижний угол

### Миниатюры
- `thumbnail=true`: создание уменьшенной копии
- Автоматическое определение размера миниатюры

## Структура хранения

- **Исходные изображения**: хранятся в `storage/originals/`
- **Обработанные изображения**: хранятся в `storage/processed/`
- **Метаданные**: хранится в структуре RepositoryInMemory, реализующей сервисный интерфейс Repository
- **Очередь задач**: Apache Kafka

## Статусы обработки

- `uploaded`: изображение загружено
- `processing`: в обработке
- `completed`: обработка завершена
- `failed`: ошибка обработки

## Поддержка

При возникновении проблем:

1. **Проверьте логи контейнеров**:
   ```bash
   docker compose logs image-processor
   docker compose logs kafka
   ```

2. **Убедитесь, что все сервисы запущены**:
   ```bash
   docker compose ps
   ```

3. **Проверьте доступность Kafka**:
   ```bash
   docker compose exec kafka kafka-topics --list --bootstrap-server localhost:9092
   ```

4. **Проверьте хранилище изображений**:
   ```bash
    docker compose exec minio ls -la /data/images
    docker compose exec minio ls /data/images/processed
   ```

## Форматы поддержки
- **Входные форматы**: JPEG, PNG, GIF
- **Выходные форматы**: JPEG, PNG, GIF
- **Максимальный размер**: 10MB

## Производительность

- Обработка изображений происходит асинхронно через Kafka
- Поддержка параллельной обработки множества изображений
- Автоматическое масштабирование обработчиков (указывается через env.WORKER_COUNT)
- Кэширование готовых изображений

## Мониторинг

Для мониторинга состояния очереди Kafka можно использовать:
```bash
docker compose exec kafka kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic image-processing-tasks \
  --from-beginning
```

## Безопасность

- Валидация входных файлов
- Ограничение размера файлов
- Проверка MIME-типов


---

**Примечание**: Обработка изображений происходит в фоновом режиме. После загрузки изображение сразу возвращает ID, а статус можно проверять через GET запросы или веб-интерфейс.
**Примечание**: Есть одна ошибка, которую я поздно заметил: сервис способен менять формат изображений, но при этом сами изображения сохраняются по пути, в котором указан формат старого изображения.