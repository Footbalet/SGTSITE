FROM golang:1.24-alpine AS builder

WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение (убедитесь, что используется правильный порт)
RUN go build -o main ./cmd/api/main.go

# Финальный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем бинарный файл из builder
COPY --from=builder /app/main .

EXPOSE 8000

CMD ["./main"]