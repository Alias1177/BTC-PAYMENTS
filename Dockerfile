# Используем образ golang:1.24-alpine как базовый для этапа сборки
FROM golang:1.24-alpine AS builder

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем все файлы из текущей директории на хосте в контейнер
COPY . .

# Скачиваем зависимости, указанные в go.mod
RUN go mod download

# Компилируем Go-приложение, добавляем флаги для уменьшения размера бинарника
RUN go build -ldflags="-w -s" -o /app/server ./cmd/server

# Используем минималистичный образ alpine:latest для финального контейнера
FROM alpine:latest

# Устанавливаем сертификаты для HTTPS-запросов
RUN apk add --no-cache ca-certificates

# Копируем скомпилированный бинарник из этапа сборки
COPY --from=builder /app/server /app/server

# Копируем директорию config с конфигурационными файлами
COPY config /app/config

# Создаем группу и пользователя для повышения безопасности (не root)
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Переключаемся на пользователя appuser
USER appuser

# Открываем порт 8080 для приложения
EXPOSE 8080

# Задаем команду для запуска приложения
CMD ["/app/server"]