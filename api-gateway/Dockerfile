FROM golang:1.24.0 AS builder

WORKDIR /app

COPY go.mod ./

RUN go mod download

COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o api-gateway .

# Используем минимальный образ для финального контейнера
FROM alpine:latest

# Устанавливаем необходимые пакеты
RUN apk --no-cache add ca-certificates

# Копируем бинарник из стадии builder
WORKDIR /root/
COPY --from=builder /app/api-gateway .

# Открываем порт
EXPOSE 8080

# Запускаем приложение
CMD ["./api-gateway"]