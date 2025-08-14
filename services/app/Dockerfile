FROM golang:1.25-alpine AS builder

# Установка зависимостей для сборки
RUN apk add --no-cache git

# Создание рабочей директории
WORKDIR /app

# Копирование файлов модуля и загрузка зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копирование исходного кода
COPY . .

# Сборка приложения
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Финальный образ
FROM alpine:latest

# Установка зависимостей
RUN apk --no-cache add ca-certificates tzdata

# Создание пользователя и рабочей директории
RUN adduser -D -s /bin/sh appuser
WORKDIR /app

# Копирование скомпилированного приложения и миграций
COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations/

# Смена владельца файлов
RUN chown -R appuser:appuser /app
USER appuser

# Установка переменных окружения
ENV DB_HOST=database
ENV DB_PORT=5432
ENV DB_USER=app_user
ENV DB_PASSWORD=app_password
ENV DB_NAME=infrastructure_training
ENV PORT=8080

# Экспорт порта
EXPOSE 8080

# Запуск приложения
CMD ["./main"]