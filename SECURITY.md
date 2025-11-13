# SECURITY - Aggressive hardening applied

This automated refactor applied:
- Structural separation: cmd/, internal/app/, pkg/
- Package-level refactor: controllers, service, repository, models.
- JWT auth (github.com/golang-jwt/jwt/v5) with HS256. Replace secret in env.
- Rate limiting (in-memory token bucket) — use Redis for distributed enforcement.
- Secure headers, basic CORS considerations, CSRF comment placeholders.
- Keep secrets out of repo; use environment variables or secret manager.
- Use prepared statements in repository layer (review manually).
- Run static analysis: go vet, staticcheck, gosec before production.
