FROM golang:1.24-alpine AS builder

RUN apk add --no-cache gcc musl-dev enet-dev

WORKDIR /app

# Копируем исходный код
COPY server.go .
COPY playerControl.go .
COPY roomControl.go .
COPY serverENET.go .
COPY go.mod go.sum ./

RUN go mod download
RUN go build -o server .

# Финальный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates enet

WORKDIR /root/

COPY --from=builder /app/server .

EXPOSE 8085 7777

CMD ["./server"]