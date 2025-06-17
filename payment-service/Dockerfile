FROM golang:1.24.0 AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы модулей
COPY go.mod ./

# Скачиваем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o payment-service ./cmd/main.go

# Финальная стадия
FROM alpine:latest

# Устанавливаем необходимые пакеты
RUN apk --no-cache add ca-certificates

# Копируем бинарник
WORKDIR /root/
COPY --from=builder /app/payment-service .

# Открываем порт
EXPOSE 8082

# Запускаем сервис
CMD ["./payment-service"]