FROM golang:1.25-alpine AS builder
WORKDIR /app
RUN apk add --no-cache gcc musl-dev sqlite-dev ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o main ./cmd

FROM alpine:3.20
RUN apk add --no-cache sqlite ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
RUN mkdir -p db && chmod 777 db
EXPOSE 8080
CMD ["tail", "-f", "/dev/null"]