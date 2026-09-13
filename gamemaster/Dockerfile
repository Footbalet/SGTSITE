FROM golang:1.24-bookworm AS builder

RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc libenet-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY server.go .
COPY playerControl.go .
COPY roomControl.go .
COPY TURN.go .
COPY go.mod go.sum ./

RUN go mod download
RUN go build -o server .

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates libenet7 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /root/

COPY --from=builder /app/server .

COPY lfl.x86_64 .
RUN chmod +x lfl.x86_64

EXPOSE 8085 7777 3478

CMD ["./server"]