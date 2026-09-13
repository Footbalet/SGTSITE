FROM golang:1.24-alpine AS builder

RUN apk add --no-cache gcc musl-dev enet-dev

WORKDIR /app

# Копируем исходный код
COPY server.go .
COPY playerControl.go .
COPY roomControl.go .
COPY TURN.go .
COPY go.mod go.sum ./

RUN go mod download
RUN go build -o server .

# Финальный образ
FROM alpine:latest

# Устанавливаем enet, сертификаты и gcompat для запуска glibc-бинарников
RUN apk --no-cache add ca-certificates enet gcompat

WORKDIR /root/

# Копируем скомпилированный сервер
COPY --from=builder /app/server .

# Копируем ваш файл lfl.x86_64 из локальной папки в контейнер
COPY lfl.x86_64 .

# Даем права на исполнение (на всякий случай)
RUN chmod +x lfl.x86_64

EXPOSE 8085 7777 3478

CMD ["./server"]