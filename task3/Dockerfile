FROM golang:1.24.4-alpine

WORKDIR /app

# Установка зависимостей
RUN apk add --no-cache git postgresql-client

# Копируем сначала только mod файлы для кэширования зависимостей
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем остальные файлы проекта
COPY . .

# Собираем приложение
RUN go build -o commentTree ./cmd/main.go

EXPOSE 8080

CMD ["./commentTree"]