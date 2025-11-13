#!/usr/bin/env bash
set -e
# run_local.sh — run the app locally (without docker), connecting to a local PostgreSQL instance.
# Usage:
#   cp .env.example .env   # edit values if needed
#   ./run_local.sh         # runs migrations then starts app
#
# Requirements:
#  - Go toolchain (go 1.16+)
#  - psql (Postgres client) in PATH for migrations, or skip migrations
#  - .env file or environment variables set

# load .env if exists
if [ -f .env ]; then
  export $(grep -v '^#' .env | xargs)
fi

# helper: build or run directly
echo "Running migrations (if psql available)..."
if command -v psql >/dev/null 2>&1; then
  ./migrate.sh
else
  echo "psql not found — skipping migrations. Install psql to run migrations automatically."
fi

echo "Starting Go app locally..."
# prefer running built binary if exists, otherwise `go run .`
if [ -f ./app ]; then
  ./app
else
  go run .
fi
