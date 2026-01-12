FROM golang:1.24.0 AS builder

WORKDIR /app

# Add build arg for Go proxy
ARG GOPROXY=https://proxy.golang.org,direct

# Установка зависимостей для компиляции sqlite3 с CGO
RUN apt-get update && apt-get install -y gcc sqlite3 libsqlite3-dev
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o url-shortener ./cmd/url-shortener/main.go

FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/url-shortener ./url-shortener
COPY --from=builder /app/config/prod-docker.yaml ./config/prod-docker.yaml

EXPOSE 44044
CMD ["./url-shortener"]

