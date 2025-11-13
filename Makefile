SHELL := /usr/bin/env bash

.PHONY: run build tidy test lint clean

run:
	APP_ENV=development PORT=8080 JWT_SECRET=devsecret CGO_ENABLED=1 go run ./cmd/app

build:
	CGO_ENABLED=1 go build -ldflags="-w -s" -o bin/server ./cmd/app

test:
	go test -v -race -coverprofile=coverage.out ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/ coverage.out

tidy:
	go mod tidy


