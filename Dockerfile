FROM golang:1.25 AS builder
WORKDIR /app
RUN apt-get update && apt-get install -y gcc libsqlite3-dev
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -ldflags="-s -w" -o main ./cmd

FROM debian:stable-slim
RUN apt-get update && apt-get install -y \
    sqlite3 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /root/
COPY --from=builder /app/main .
RUN mkdir -p db && chmod 777 db
EXPOSE 8080
CMD ["tail", "-f", "/dev/null"]