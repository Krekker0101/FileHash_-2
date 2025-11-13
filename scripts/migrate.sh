#!/usr/bin/env bash
set -e

# Migration script for PostgreSQL
# Usage: ./scripts/migrate.sh

if [ -f .env ]; then
  export $(grep -v '^#' .env | xargs)
fi

DB_TYPE=${DB_TYPE:-sqlite}

if [ "$DB_TYPE" = "postgres" ]; then
  DATABASE_URL=${DATABASE_URL:-postgres://filehash:filehash@localhost:5432/filehash?sslmode=disable}
  
  echo "Running GORM migrations for PostgreSQL..."
  go run ./cmd/app migrate
else
  echo "SQLite migrations are handled automatically by GORM AutoMigrate"
fi

