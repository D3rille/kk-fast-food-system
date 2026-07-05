# NextGen Kiosk-to-Kitchen Fast Food System

This repository contains the backend and foundation code for the NextGen Kiosk-to-Kitchen Fast Food System.

## Architecture

The system is built around a local-first, cloud-second architecture to ensure that core operations function over the Local Area Network (LAN) regardless of internet connectivity.

```text
┌──────────────────────────────────────────────────────────────┐
│                      LOCAL NETWORK (LAN)                     │
│                                                              │
│  ┌─────────────┐   ┌─────────────┐   ┌──────────────────┐   │
│  │   Kiosk UI  │   │   KDS UI    │   │  Admin Dashboard │   │
│  │   (React)   │   │   (React)   │   │     (React)      │   │
│  └──────┬──────┘   └──────┬──────┘   └────────┬─────────┘   │
│         │                 │                    │             │
│         │    HTTP/JSON    │    WebSocket       │  HTTP/JSON  │
│         └────────┬────────┴────────┬───────────┘             │
│                  │                 │                          │
│         ┌────────▼─────────────────▼────────┐                │
│         │         Go API Gateway            │                │
│         │    (Chi Router + WebSocket Hub)    │                │
│         └────────┬─────────────────┬────────┘                │
│                  │     gRPC        │                          │
│         ┌────────▼────────┐ ┌──────▼──────────┐              │
│         │  Order Service  │ │  Menu Service   │              │
│         │  Auth Service   │ │  Kitchen Service│              │
│         └────────┬────────┘ └──────┬──────────┘              │
│                  │                 │                          │
│         ┌────────▼─────────────────▼────────┐                │
│         │           PostgreSQL              │                │
│         └───────────────────────────────────┘                │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

For full details, please refer to:
- [Business Requirements Document (BRD.md)](documentation/BRD.md)
- [Technical Implementation Plan (PlanA.md)](documentation/PlanA.md)

## Tech Stack
- **Backend:** Go 1.26+, Chi v5 Router, pgx v5 PostgreSQL Driver, slog logging, Viper configurations.
- **Database:** PostgreSQL 17.
- **Deployment:** Docker & Docker Compose.

## Prerequisites
- Go 1.26.4 or higher
- PostgreSQL 17
- Docker & Docker Compose
- `golangci-lint` (for linting)
- `golang-migrate` (for migrations)

## Get Started

### Local Config
1. Copy the example environment file:
   ```bash
   cp configs/.env.example configs/.env
   ```
2. Adjust the variables in `configs/.env` as needed.

### Running locally
To build and run the Go API server locally:
```bash
make build
make run
```

### Docker Compose
To run PostgreSQL and the backend API service inside docker:
```bash
make docker-up
```

### Tests & Linter
- Run tests: `make test`
- Run lint check: `make lint`
