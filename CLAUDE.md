# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Run all tests (requires Docker for store integration tests)
go test ./...

# Run a single test
go test ./internal/handler/ -run TestListHandler_GetLists -v

# Run only unit tests (skip testcontainers)
go test -short ./...

# Run the server locally (requires .env)
go run ./cmd/server

# Apply DB migrations manually
migrate -path migrations -database "$DATABASE_URL" up

# Build
go build -o todo-service ./cmd/server

# Start local stack (postgres + service)
docker compose up -d --build
```

## Architecture

Single-binary Go service. All application code lives under `internal/` and is not importable by external packages.

**Request flow:**
```
chi router
  → middleware.CORS       (global, handles OPTIONS preflights)
  → middleware.MaxBodySize (global, 1 MB limit)
  → middleware.Auth        (authenticated routes only)
      reads httpOnly cookie "token" → parses HS256 JWT → stores sub claim
      as userID in request context via UserIDFromContext()
  → handler.{List,Todo}Handler
  → store.{PgListStore,PgTodoStore}
  → PostgreSQL
```

**Key design decisions:**
- Ownership checks always return **404** (not 403) to avoid leaking resource existence to other users.
- All handlers assert `UserIDFromContext` returns `ok=true`; false means a route was registered outside the auth group (programmer error → 500).
- `UpdateTodo` returns 400 if both `text` and `done` are absent — empty PATCHes are rejected.
- Store interfaces (`ListStore`, `TodoStore`) are defined in `store/store.go`; implementations in `list_store.go` / `todo_store.go`.
- Migrations are embedded into the binary via `//go:embed *.sql` in `migrations/embed.go` and run automatically on startup using `golang-migrate` + `iofs` driver.
- JWT secret is standard Base64 (`base64.StdEncoding`), not URL-safe — must match the encoding used by `auth.savuliak.com`.

## Environment

Copy `.env.example` to `.env`. Required vars: `DATABASE_URL`, `JWT_SECRET` (standard Base64, ≥32 bytes decoded, same secret as auth service). Optional: `PORT` (default `8080`), `COOKIE_NAME` (default `token`), `CORS_ALLOWED_ORIGINS` (comma-separated, default `http://localhost:5173`).

Local postgres is exposed on **port 5434** (5433 is taken by user-service).

## Deploy

CI runs `go mod tidy` then `go test ./...`, then rsync to `/home/deploy/todo-service` on `savuliak.com` and runs:
```
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d --build
```
The `.env` file must be pre-provisioned on the server — it is excluded from rsync. Required GitHub secrets: `DO_KEY_CI`, `DO_HOST`, `DO_USER`.

The service joins the existing `gateway-network` (nginx-proxy + acme-companion) and is served at `todo.savuliak.com`.
