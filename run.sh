#!/bin/bash
export CGO_ENABLED=1
export APP_ENV=development
export PORT=8080
export JWT_SECRET=devsecret
go run ./cmd/app

