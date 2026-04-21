FROM golang:1.24-alpine AS builder

WORKDIR /app

# Устанавливаем зависимости для компиляции ENet
RUN apk add --no-cache gcc musl-dev linux-headers

# Копируем исходный код
COPY server.go .
COPY playerControl.go .
COPY roomControl.go .
COPY serverENET.go .
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Компилируем с поддержкой статической линковки
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o server .

# Финальный образ
FROM alpine:latest

# Устанавливаем необходимые библиотеки для работы ENet
RUN apk --no-cache add ca-certificates libc6-compat

WORKDIR /root/

# Копируем бинарник
COPY --from=builder /app/server .

# Открываем порты
# 7777 - ENet UDP порт (основной)
# 8085 - возможно HTTP или другой порт (если нужен)
EXPOSE 7777/udp 8085/tcp

# Запускаем сервер
CMD ["./server"]