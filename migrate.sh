#!/usr/bin/env bash
set -e
# migrate.sh — apply SQL files from ./migrations in alphabetical order using psql.
# Environment variables used:
#   PG_HOST, PG_PORT, PG_USER, PG_PASSWORD, PG_DB
# or DATABASE_URL can be used for psql connection string.
# Example:
#   DATABASE_URL=postgres://user:pass@host:5432/dbname?sslmode=disable ./migrate.sh

# load .env if exists
if [ -f .env ]; then
  export $(grep -v '^#' .env | xargs)
fi

CONN=""
if [ -n "$DATABASE_URL" ]; then
  CONN="$DATABASE_URL"
else
  : "${PG_HOST:=127.0.0.1}"
  : "${PG_PORT:=5432}"
  : "${PG_USER:=postgres}"
  : "${PG_PASSWORD:=postgres}"
  : "${PG_DB:=postgres}"
  CONN="postgresql://$PG_USER:$PG_PASSWORD@$PG_HOST:$PG_PORT/$PG_DB"
fi

echo "Using connection: $CONN"

# psql needs environment PGPASSWORD or a connection string with password
export PGPASSWORD="${PG_PASSWORD:-$(echo $CONN | sed -n 's/.*:\/\///;s/.*:\(.*\)@.*/\1/p')}"
# Apply migrations
if [ -d migrations ]; then
  for f in $(ls migrations/*.sql 2>/dev/null | sort); do
    echo "Applying $f"
    psql "$CONN" -f "$f"
  done
else
  echo "No migrations directory found. Skipping."
fi
