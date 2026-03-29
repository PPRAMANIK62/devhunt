# Stage 1: Build
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy dependency files first - Docker caches this layer.
# Only re-downloads modules if go.mod/go.sum changes.
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o devhunt ./cmd/api

# Stage 2: Run
# Alpine is ~5MB vs ~300MB for the Go image
FROM alpine:3.19

WORKDIR /app

# ca-certificates needed for HTTPS calls to external services
RUN apk --no-cache add ca-certificates

COPY --from=builder /app/devhunt .
COPY --from=builder /app/internal/database/migrations ./internal/database/migrations

EXPOSE 8080

CMD ["./devhunt"]
