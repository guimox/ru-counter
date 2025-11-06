FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache gcc musl-dev sqlite-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=1
ENV GOOS=linux
RUN cd cmd && go build -a -ldflags '-linkmode external -extldflags "-static"' -o main .

FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite

WORKDIR /root/

COPY --from=builder /app/cmd/main .

RUN mkdir -p db

EXPOSE 8080

CMD ["tail", "-f", "/dev/null"]