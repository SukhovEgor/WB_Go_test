# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Копируем модули первыми — для кэширования
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь проект
COPY . .

# Собираем из правильной директории (там где main.go)
# Предполагаем, что main.go находится в scripts/kafka
RUN CGO_ENABLED=0 GOOS=linux go build -o /kafka ./scripts/kafka

# Даём права на выполнение — обязательно!
RUN chmod +x /kafka

# Final stage
FROM alpine:latest

# Добавляем сертификаты (нужно для HTTPS, если используете)
RUN apk add --no-cache ca-certificates

# Копируем бинарник
COPY --from=builder /kafka /kafka

# Рабочая директория (опционально)
WORKDIR /app

# Порт, который слушает HTTP-сервер
EXPOSE 8082

# Запуск
CMD ["/kafka"]