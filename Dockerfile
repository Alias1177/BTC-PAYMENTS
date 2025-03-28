# Используем образ Golang для сборки приложения
FROM golang:1.24-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем все файлы в контейнер
COPY . .

# Загружаем зависимости
RUN go mod download

# Собираем приложение
RUN go build -ldflags="-w -s" -o /app/service ./cmd/service

# Используем минимальный образ для запуска приложения
FROM alpine:latest

# Устанавливаем необходимые пакеты
RUN apk add --no-cache ca-certificates

# Копируем собранное приложение из предыдущего этапа
COPY --from=builder /app/service /app/service

# Копируем конфигурацию
COPY config /app/config

# Создаем группу и пользователя для запуска приложения
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Переключаемся на созданного пользователя
USER appuser

# Открываем порт для приложения
EXPOSE 8080

# Запускаем приложение
CMD ["/app/service"]