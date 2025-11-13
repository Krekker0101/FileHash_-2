FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache build-base git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build application
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o server ./cmd/app

FROM alpine:3.20

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -S app && adduser -S app -G app

# Copy binary from builder
COPY --from=builder /app/server /app/server

# Create necessary directories
RUN mkdir -p /app/uploads /app/data && \
    chown -R app:app /app

USER app

# Environment variables
ENV APP_ENV=production
ENV HTTP_HOST=0.0.0.0
ENV HTTP_PORT=8080
ENV UPLOADS_DIR=/app/uploads
ENV DATABASE_PATH=/app/data/filehash.db

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/healthz || exit 1

ENTRYPOINT ["/app/server"]

