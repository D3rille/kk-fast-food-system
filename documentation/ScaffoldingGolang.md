For a **production-ready Go project**, the fastest approach is usually **not** starting from an empty directory. Instead, begin with a well-structured template and strip away what you don't need.

## Option 1: Start Simple (Recommended for Most Projects)

For APIs and backend services, a structure like this scales surprisingly well:

```text
myapp/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── handlers/
│   ├── service/
│   ├── repository/
│   └── config/
├── pkg/
├── migrations/
├── configs/
├── scripts/
├── tests/
├── .env.example
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── go.mod
└── README.md
```

Initialize:

```bash
mkdir myapp
cd myapp

go mod init github.com/yourname/myapp

mkdir -p cmd/api
mkdir -p internal/{handlers,service,repository,config}
mkdir migrations scripts tests configs
```

## Option 2: Use a Production Template

A popular starting point is the Go project layout maintained by the community:

[golang-standards/project-layout](https://github.com/golang-standards/project-layout?utm_source=chatgpt.com)

Be aware that many Go maintainers consider it a reference rather than an official standard. Don't copy everything blindly.

## Stack I Usually Recommend

### HTTP Framework

For performance and simplicity:

* [Chi](https://github.com/go-chi/chi?utm_source=chatgpt.com)
* [Gin](https://gin-gonic.com?utm_source=chatgpt.com)
* Standard `net/http`

Today I'd lean toward **Chi** or plain `net/http`.

### Configuration

```bash
go get github.com/spf13/viper
```

or simply:

```bash
go get github.com/caarlos0/env/v11
```

Many Go teams now prefer environment variables with `caarlos0/env` over large config frameworks.

### Logging

```bash
go get go.uber.org/zap
```

or use Go's built-in:

```go
log/slog
```

Since Go 1.21+, **slog** is usually sufficient.

### Database

PostgreSQL +:

```bash
go get github.com/jackc/pgx/v5
```

Avoid ORMs initially unless your team specifically wants one.

### Migrations

```bash
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### Dependency Injection

For most projects:

```go
func NewUserService(repo UserRepository) *UserService
```

Manual DI is enough.

For larger systems:

* [Google Wire](https://github.com/google/wire?utm_source=chatgpt.com)

## Add Production Essentials Immediately

### Graceful Shutdown

```go
ctx, stop := signal.NotifyContext(
    context.Background(),
    os.Interrupt,
    syscall.SIGTERM,
)
defer stop()
```

### Health Checks

```text
GET /healthz
GET /readyz
```

### Structured Logging

```go
logger.Info("user created",
    "user_id", user.ID,
)
```

### Timeouts

```go
srv := &http.Server{
    Addr: ":8080",
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  30 * time.Second,
}
```

### Docker

```dockerfile
FROM golang:1.25 AS builder

WORKDIR /app

COPY . .
RUN CGO_ENABLED=0 go build -o server ./cmd/api

FROM gcr.io/distroless/static-debian12

COPY --from=builder /app/server /server

ENTRYPOINT ["/server"]
```

## My "Fast Production API" Starter Checklist

```text
✓ Go modules
✓ Chi router
✓ PostgreSQL + pgx
✓ slog logging
✓ env-based config
✓ migrations
✓ Dockerfile
✓ Makefile
✓ graceful shutdown
✓ health endpoints
✓ GitHub Actions CI
✓ tests
✓ linting
```

### Add these tools on day one

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

```bash
go test ./...
golangci-lint run
```

## If You Want Maximum Speed

A lot of Go developers today scaffold an API with:

* Chi
* pgx
* slog
* Docker
* golangci-lint
* GitHub Actions

and let an AI assistant generate the repetitive repository/handler/service boilerplate.

That gives you a production-ready foundation in under an hour without introducing a large framework that you'll later fight against.
