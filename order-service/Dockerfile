# Build stage
FROM golang:1.24.0 AS builder

WORKDIR /app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o order-service ./cmd/main.go

# Runtime stage
FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/order-service .

EXPOSE 8080
CMD ["./order-service"]