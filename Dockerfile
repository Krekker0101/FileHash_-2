FROM golang:1.24-alpine AS builder
WORKDIR /app
RUN apk add --no-cache build-base
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o server ./cmd/app

FROM alpine:3.20
WORKDIR /app
RUN addgroup -S app && adduser -S app -G app
USER app
COPY --from=builder /app/server /app/server
COPY migrations /app/migrations
ENV APP_ENV=production PORT=8080 UPLOADS_DIR=/app/uploads DATABASE_PATH=/app/data/filehash.db
EXPOSE 8080
ENTRYPOINT ["/app/server"]


