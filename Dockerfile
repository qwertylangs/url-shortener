FROM golang:1.24.0-bookworm AS builder
# glibc а не musl как в alpine (для работы с sqlite3)

WORKDIR /app

# Add build arg for Go proxy
ARG GOPROXY=https://proxy.golang.org,direct

# Установка зависимостей для компиляции sqlite3 с CGO
RUN apt-get update && apt-get install -y gcc sqlite3 libsqlite3-dev
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o url-shortener ./cmd/url-shortener/main.go
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o migrator ./cmd/migrator/main.go

FROM debian:bookworm-slim
WORKDIR /app
# Устанавливаем библиотеки для работы SQLite в рантайме wget для health check
RUN apt-get update && apt-get install -y libsqlite3-0 ca-certificates curl && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/url-shortener ./url-shortener
COPY --from=builder /app/migrator ./migrator
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/config/prod-docker.yaml ./config/prod-docker.yaml

EXPOSE 8082
CMD ["./url-shortener"]

