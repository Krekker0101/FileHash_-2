# How to run this project (automated additions)

Detected primary language: **go**

I added Docker configuration and helper scripts so you can run the project reliably.

Common commands (from project root):

1. With Docker (recommended):
   - `docker-compose up --build` — builds images and runs the app.
   - Open http://localhost:8080 (or check docker-compose ports).

2. Without Docker:
   - See language-specific notes below.

Language-specific notes:

- This project appears to be Go-based.
- I added a Dockerfile that builds the Go binary and runs it.
- The Dockerfile assumes `main.go` or a `cmd/` entrypoint. Adjust as needed.


## PostgreSQL (local and Docker) — how to run

### Option A — Docker (recommended for reproducibility)
1. Build and run both app + postgres:
   ```bash
   docker-compose up --build
   ```
   - The `db` service uses Postgres 15 and initialises the database using SQL files placed in `./migrations`.
   - App will be reachable on port 8080 (http://localhost:8080) if it listens on that port.

2. To run only the database:
   ```bash
   docker-compose up -d db
   ```
   Then run the app locally (see Option B) connecting to `127.0.0.1:5432` (or to `db:5432` from within Docker).

### Option B — Local Postgres (no Docker)
1. Install PostgreSQL locally (your OS package manager).
2. Create DB / user (example):
   ```bash
   sudo -u postgres psql -c "CREATE USER postgres WITH PASSWORD 'postgres';"
   sudo -u postgres psql -c "CREATE DATABASE postgres OWNER postgres;"
   ```
3. Copy `.env.example` to `.env` and edit values if needed:
   ```bash
   cp .env.example .env
   # edit .env
   ```
4. Run migrations:
   ```bash
   ./migrate.sh
   ```
5. Run the app:
   ```bash
   ./run_local.sh
   ```
   or build first:
   ```bash
   go build -o app .
   ./app
   ```

Notes:
- `migrate.sh` applies SQL files from the `migrations/` directory in alphabetical order using `psql`.
- For Docker convenience, `migrations/` is mounted into the Postgres container init directory so initial data/schema will be applied automatically.
- If your Go app expects a different env variable names, adjust `.env.example` and the code accordingly.
