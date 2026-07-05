# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make build          # compile binary to bin/server
make run            # go run ./cmd/api
make test           # go test -v -race -count=1 ./...
make lint           # golangci-lint run ./...
make vet            # go vet ./...
make fmt            # go fmt ./...

make migrate-up     # apply migrations (uses DATABASE_URL env or Makefile default)
make migrate-down   # revert migrations
make migrate-create name=<migration_name>  # create a new numbered migration file

make docker-up      # build + start all services via Docker Compose
make docker-down    # stop and remove containers
```

Run a single test package: `go test -v -race ./internal/service/...`

## Environment

Copy `configs/.env.example` to `configs/.env` before running locally. Config is loaded via Viper from that file.

## Architecture

The project follows a strict layered architecture wired together in `cmd/api/main.go`:

```
Handler → Service (interface) → Repository (interface) → pgxpool.Pool (PostgreSQL)
```

- **`internal/config/`** — Viper-based config, reads `configs/.env`
- **`internal/database/`** — `pgxpool.Pool` connection setup
- **`internal/models/`** — domain structs and enums; DTOs (request/response shapes) are in separate `*_dto.go` files alongside each domain
- **`internal/repository/`** — raw SQL via pgx; each repository defines an interface that the service layer depends on
- **`internal/service/`** — business logic; depends only on repository interfaces, never on `pgxpool` directly
- **`internal/handlers/`** — Chi HTTP handlers; depend on service interfaces
- **`internal/middleware/`** — Chi middlewares: Recovery, Logging (slog), CORS, `Authenticate` (JWT), `RequireRole`
- **`migrations/`** — sequential SQL files managed by `golang-migrate`

## Key conventions

**Monetary values** are stored and passed as `int64` in centavos (Philippine pesos × 100), never as floats.

**JWT auth** uses two token types (`"access"` / `"refresh"`). The `Authenticate` middleware (`internal/middleware/auth.go`) enforces access-token-only on protected routes and injects claims into context. Retrieve them with `middleware.GetUserID(ctx)` and `middleware.GetUserRole(ctx)`.

**Role-based access** is applied by chaining `middleware.RequireRole(...)` after `middleware.Authenticate(...)`. Admin/manager routes live under `/api/v1/admin/`. Currently implemented roles: `admin`, `manager`, `cashier`, `kitchen`.

**Repositories** accept the domain struct as a parameter (not individual fields) for Create/Update. SQL is written inline with `$N` positional parameters (pgx style, not `?`).

**IDs** are UUID strings. Auth-related UUIDs use a custom generator in `service/auth.go`; order IDs use `github.com/google/uuid` v7 (`uuid.NewV7()`) for time-ordered sorting.

## Linting

golangci-lint is configured in `.golangci.yml` with: `govet`, `errcheck`, `staticcheck`, `unused`, `gosimple`, `ineffassign`, `gocritic`, `gofmt`, `goimports`. The `errcheck` linter is suppressed in `_test.go` files.


## Developer Guide
To understand **how** the project is structured, **why** it was designed this way, and **how** you can safely add new features, APIs, and services without breaking the architecture, refer to [Developer Guide](./documentation/developer_guide.md). Before implementing any feature or refactoring, read the architecture rules in this guide using the file read tool.

## Business Requirements Document
Features to be built must adhere and aim to accomplish accomplish the [BRD](./documentation/BRD.md).