# Build stage
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server ./cmd/server/

# Runtime stage
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata curl
ENV TZ=Asia/Tehran

WORKDIR /app

COPY --from=builder /app/server .
# Migrations run separately via golang-migrate / Flyway — not baked into image

# Create non-root user
RUN adduser -D -g '' appuser

# Create keys directory with write permission for JWT key generation
RUN mkdir -p /app/keys && chown appuser:appuser /app/keys
USER appuser

EXPOSE 8080

HEALTHCHECK --interval=15s --timeout=3s --retries=3 \
  CMD wget -q --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["./server"]
