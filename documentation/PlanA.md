# Technical Implementation Plan (Plan A)

**Project Name:** NextGen Kiosk-to-Kitchen Fast Food System

**Document Version:** 2.0

**Date:** June 9, 2026

**Author:** Engineering Lead

**Companion Document:** [BRD.md](file:///c:/Users/Djhanggoo/Documents/Projects/FastFood%20System/fastFoodSystem/documentation/BRD.md) — Business Requirements Document

---

## 1. Overview

This document defines the **how** — the technology choices, system architecture, service design, deployment strategy, and implementation roadmap for the NextGen Kiosk-to-Kitchen Fast Food System.

All decisions in this plan directly trace back to the business requirements defined in [BRD.md](file:///c:/Users/Djhanggoo/Documents/Projects/FastFood%20System/fastFoodSystem/documentation/BRD.md). The two guiding architectural principles are:

1. **Local-first, cloud-second** — all core operations run on a LAN; cloud sync is asynchronous.
2. **Multi-channel ready** — the backend is designed as if online and mobile ordering already exist, even though Phase 1 only implements kiosk and cashier channels.

---

## 2. System Architecture

### 2.1 High-Level Architecture

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
                           │
                    (async cloud sync)
                           │
                    ┌──────▼──────┐
                    │  Cloud DB   │
                    │  (Phase 2+) │
                    └─────────────┘
```

### 2.2 Communication Patterns

| Path                    | Protocol       | Rationale                                              |
|-------------------------|----------------|--------------------------------------------------------|
| Frontend → API Gateway  | HTTP/JSON      | Simple, browser-native, easy to debug                  |
| KDS → API Gateway       | WebSocket      | Real-time push for instant order display (BRD KDS-001) |
| API Gateway → Services  | gRPC           | Typed contracts, fast serialization, easy future scaling|
| Services → Database     | pgx (TCP)      | Direct PostgreSQL driver, no ORM overhead               |

> **Design Decision:** gRPC is used for internal service-to-service communication only. It is never exposed directly to browser clients. The API Gateway translates between HTTP/JSON (external) and gRPC (internal).

### 2.3 Multi-Channel Order Source

```go
type OrderSource string

const (
    KioskSource   OrderSource = "kiosk"
    CashierSource OrderSource = "cashier"
    MobileSource  OrderSource = "mobile"  // Phase 2+
    OnlineSource  OrderSource = "online"  // Phase 2+
)
```

The kitchen is source-agnostic — orders from all channels appear identically on the KDS with a source badge (BRD KDS-005).

---

## 3. Technology Stack

### 3.1 Backend

| Component       | Technology           | Rationale                                                    |
|-----------------|----------------------|--------------------------------------------------------------|
| Language        | **Go 1.26+**         | Fast startup, small memory footprint, strong concurrency     |
| HTTP Router     | **Chi v5**           | Lightweight, idiomatic, middleware-friendly                  |
| Database Driver | **pgx v5**           | High-performance PostgreSQL driver, no ORM overhead          |
| Logging         | **slog** (stdlib)    | Structured logging, sufficient since Go 1.21+                |
| Configuration   | **Viper**            | Environment + file-based config (already in go.mod)          |
| Migrations      | **golang-migrate**   | SQL-based, version-controlled schema migrations              |
| Internal RPC    | **gRPC**             | Typed service contracts, fast binary serialization           |
| Real-time Push  | **gorilla/websocket**| WebSocket hub for KDS real-time order feed                   |
| Auth            | **JWT (go-jwt)**     | Stateless token-based authentication                         |

### 3.2 Frontend

| Component       | Technology           | Rationale                                                    |
|-----------------|----------------------|--------------------------------------------------------------|
| Framework       | **React 19**         | Component model, huge ecosystem, easy hiring                 |
| Language        | **TypeScript**       | Type safety, better developer experience                     |
| Styling         | **Tailwind CSS**     | Rapid prototyping, consistent design tokens                  |
| Data Fetching   | **TanStack Query**   | Caching, refetching, WebSocket integration                   |
| Build Tool      | **Vite**             | Fast HMR, optimized production builds                        |

### 3.3 Infrastructure

| Component       | Technology           | Rationale                                                    |
|-----------------|----------------------|--------------------------------------------------------------|
| Containerization| **Docker**           | Reproducible builds, easy deployment                         |
| Orchestration   | **Docker Compose**   | Single-command deployment, no K8s complexity for Phase 1     |
| Database        | **PostgreSQL 17**    | ACID compliance, JSONB for flexible data, battle-tested      |
| Hardware        | **Mini PC / Local Server** | Local-first deployment per BRD Section 5.2              |

---

## 4. Project Structure

```text
fastFoodSystem/
├── cmd/
│   └── api/
│       └── main.go                 # Application entry point
├── internal/
│   ├── config/                     # Configuration loading (Viper)
│   ├── handlers/                   # HTTP request handlers (Chi routes)
│   ├── service/                    # Business logic layer
│   ├── repository/                 # Database access layer (pgx)
│   ├── models/                     # Domain entities & DTOs
│   ├── middleware/                 # Auth, logging, CORS middleware
│   └── ws/                        # WebSocket hub for KDS real-time
├── proto/                          # gRPC protobuf definitions
├── migrations/                     # SQL migration files
├── configs/                        # Config files (.env, YAML)
├── scripts/                        # Build, deploy, seed scripts
├── tests/                          # Integration & E2E tests
├── documentation/
│   ├── BRD.md                      # Business Requirements Document
│   ├── PlanA.md                    # This document
│   └── ScaffoldingGolang.md        # Go project scaffolding guide
├── web/                            # Frontend applications
│   ├── kiosk/                      # Kiosk UI (React)
│   ├── kds/                        # Kitchen Display UI (React)
│   └── admin/                      # Admin Dashboard (React)
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── go.mod
├── go.sum
└── README.md
```

---

## 5. Service Design

### 5.1 Order Service

Implements the core order state machine defined in BRD Section 3.1:

```text
[Draft] → [Pending Payment] → [Paid/Queued] → [In Preparation] → [Ready for Pickup] → [Completed]
                │                                                                          │
                └─► [Cancelled]                                                [Cancelled] ◄┘
```

**Responsibilities:**
- Create, update, and cancel orders
- Manage order lifecycle transitions with validation
- Emit WebSocket events on state changes (for KDS real-time feed)
- Record order source channel (kiosk, cashier, etc.)

### 5.2 Menu Service

**Responsibilities:**
- CRUD for products, categories, modifier groups, and modifier options
- Real-time availability toggling (BRD AM-001: kill-switch)
- Flexible product/modifier schema supporting combos, sizes, add-ons

### 5.3 Kitchen Service

**Responsibilities:**
- Manage KDS ticket queue (FIFO ordering)
- Calculate ticket aging and color-coding (BRD KDS-002)
- Handle ticket state transitions (In Prep → Ready for Pickup)
- WebSocket broadcast to all connected KDS screens

### 5.4 Auth Service

**Responsibilities:**
- User registration and login (staff only for Phase 1)
- JWT token generation and validation
- Role-based access control enforcement

### 5.5 Payment Service

**Responsibilities:**
- Abstracted payment provider interface

```go
type PaymentProvider interface {
    CreatePayment(ctx context.Context, req CreatePaymentRequest) (*PaymentResult, error)
    VerifyPayment(ctx context.Context, transactionRef string) (*PaymentStatus, error)
    RefundPayment(ctx context.Context, transactionRef string, amount int64) error
}
```

- Phase 1 implementations: `CashProvider`, `CardTerminalStubProvider`
- Phase 2+ implementations: `GCashProvider`, `MayaProvider`, `PayMongoProvider`, `StripeProvider`

---

## 6. API Design

### 6.1 Public Endpoints (Kiosk — No Auth)

```text
GET    /api/v1/menu/categories           # List active categories
GET    /api/v1/menu/categories/:id/items  # List available items in category
GET    /api/v1/menu/items/:id             # Get item detail with modifiers
POST   /api/v1/orders                     # Create a new order
GET    /api/v1/orders/:id/status          # Check order status (by order number)
```

### 6.2 Protected Endpoints (Staff — JWT Required)

```text
# Orders
GET    /api/v1/orders                     # List orders (filtered by status/date)
PATCH  /api/v1/orders/:id/status          # Transition order state
GET    /api/v1/orders/:id                 # Get full order detail

# Menu Management
POST   /api/v1/admin/menu/items           # Create item
PUT    /api/v1/admin/menu/items/:id       # Update item
PATCH  /api/v1/admin/menu/items/:id/availability  # Toggle availability (kill-switch)
POST   /api/v1/admin/menu/categories      # Create category
PUT    /api/v1/admin/menu/categories/:id  # Update category

# Modifier Management
POST   /api/v1/admin/menu/modifier-groups         # Create modifier group
PUT    /api/v1/admin/menu/modifier-groups/:id     # Update modifier group
POST   /api/v1/admin/menu/modifier-options        # Create modifier option
PUT    /api/v1/admin/menu/modifier-options/:id    # Update modifier option

# Auth
POST   /api/v1/auth/login                # Staff login
POST   /api/v1/auth/refresh              # Refresh JWT token

# Payments
POST   /api/v1/payments                  # Process payment for order
GET    /api/v1/payments/:id              # Get payment status
```

### 6.3 WebSocket Endpoints

```text
WS     /api/v1/ws/kds                    # KDS real-time order feed
```

**WebSocket Event Types:**

```json
{ "type": "order.created",       "data": { "order": {...} } }
{ "type": "order.status_changed", "data": { "orderId": "...", "from": "paid", "to": "in_preparation" } }
{ "type": "order.cancelled",     "data": { "orderId": "..." } }
{ "type": "menu.item_updated",   "data": { "itemId": "...", "isAvailable": false } }
```

### 6.4 Health Endpoints

```text
GET    /healthz                          # Liveness check
GET    /readyz                           # Readiness check (DB connected, etc.)
```

---

## 7. Database Schema Highlights

Based on BRD Section 6, key design decisions:

1. **StoreID on all tables** — even for single-store Phase 1, include `store_id` as a foreign key for future multi-store expansion.
2. **Order source enum** — `order_source` column on orders table: `kiosk`, `cashier`, `mobile`, `online`.
3. **Flexible modifier schema** — Products link to ModifierGroups via a join table (`product_modifier_groups`), allowing reuse of modifier groups across products.
4. **Payment table** — separate from orders, supporting multiple payment attempts per order and different providers.
5. **Soft deletes** — menu items and categories use `is_active` flags rather than hard deletes to preserve referential integrity with historical orders.

---

## 8. Production Essentials (Day One)

### 8.1 Graceful Shutdown

```go
ctx, stop := signal.NotifyContext(
    context.Background(),
    os.Interrupt,
    syscall.SIGTERM,
)
defer stop()
```

### 8.2 HTTP Server Timeouts

```go
srv := &http.Server{
    Addr:         ":8080",
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  30 * time.Second,
}
```

### 8.3 Structured Logging

```go
logger.Info("order state changed",
    "order_id", order.ID,
    "from", previousState,
    "to", newState,
    "actor", userID,
)
```

### 8.4 Docker Deployment

```dockerfile
FROM golang:1.26 AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o server ./cmd/api

FROM gcr.io/distroless/static-debian12
COPY --from=builder /app/server /server
ENTRYPOINT ["/server"]
```

```yaml
# docker-compose.yml
services:
  postgres:
    image: postgres:17
    environment:
      POSTGRES_DB: fastfood
      POSTGRES_USER: app
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - pgdata:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  backend:
    build: .
    depends_on:
      - postgres
    environment:
      DATABASE_URL: postgres://app:${DB_PASSWORD}@postgres:5432/fastfood?sslmode=disable
    ports:
      - "8080:8080"

  kiosk:
    build: ./web/kiosk
    ports:
      - "3000:80"

  kds:
    build: ./web/kds
    ports:
      - "3001:80"

  admin:
    build: ./web/admin
    ports:
      - "3002:80"

volumes:
  pgdata:
```

---

## 9. Quality & Tooling

### 9.1 Testing Strategy

| Type              | Tool               | Coverage Target                                |
|-------------------|--------------------|------------------------------------------------|
| Unit Tests        | `go test`          | Service and repository layers                  |
| Integration Tests | `go test` + testdb | API endpoints with real PostgreSQL (testcontainers)|
| Linting           | `golangci-lint`    | Enforced in CI                                 |
| Frontend Tests    | Vitest + RTL       | Component and integration tests                |

### 9.2 CI/CD Pipeline

```text
Push → Lint → Test → Build Docker Images → (Deploy to staging)
```

### 9.3 Day One Tooling Checklist

```text
✓ Go modules initialized
✓ Chi router configured
✓ PostgreSQL + pgx connected
✓ slog structured logging
✓ Viper config loading
✓ Database migrations (golang-migrate)
✓ Dockerfile (multi-stage)
✓ docker-compose.yml
✓ Makefile (build, test, migrate, lint)
✓ Graceful shutdown
✓ Health endpoints (/healthz, /readyz)
✓ golangci-lint configured
✓ Basic test suite
```

---

## 10. Implementation Roadmap

### Phase 1: MVP (Current)

```text
Sprint 1: Foundation
  ├── Database schema & migrations
  ├── Auth service (JWT + roles)
  ├── Menu service (CRUD + availability)
  └── Health endpoints

Sprint 2: Core Ordering
  ├── Order service (state machine)
  ├── Payment service (cash provider)
  ├── WebSocket hub for KDS
  └── API endpoint implementation

Sprint 3: Frontend
  ├── Kiosk UI (menu browsing, cart, checkout)
  ├── KDS UI (real-time order feed, aging, state transitions)
  └── Admin dashboard (menu management, availability toggle)

Sprint 4: Integration & Polish
  ├── End-to-end testing
  ├── Receipt/ticket printing integration
  ├── Docker Compose deployment
  └── Performance validation (BRD latency targets)
```

### Phase 2: Expansion (Future)

```text
  ├── Mobile app / online ordering channels
  ├── GCash, Maya, PayMongo, Stripe payment providers
  ├── Customer accounts & loyalty programs
  ├── Multi-store organization management
  ├── Cloud sync worker
  ├── Inventory management & analytics
  └── OIDC/SSO authentication
```

---

## 11. Offline-First Design Details

This is the most critical architectural requirement (BRD Section 5.2).

### Local-First Topology

```text
┌─────────────────── Mini PC / Local Server ────────────────────┐
│                                                                │
│   [PostgreSQL]  [Go Backend]  [React Frontend (Nginx)]        │
│                                                                │
└────────────────────────┬───────────────────────────────────────┘
                         │ LAN
          ┌──────────────┼──────────────┐
          │              │              │
     [Kiosk #1]    [Kiosk #2]    [KDS Screen]
```

### Design Rules

1. **All services run on the local network.** No external API calls are in the critical ordering path.
2. **PostgreSQL is the single source of truth locally.** No SQLite, no in-memory-only stores.
3. **Cloud sync is asynchronous and non-blocking.** A background worker queues local changes and replays them to the cloud database when internet is available.
4. **Failure mode:** If the cloud is unreachable, the system continues operating indefinitely on LAN. No degradation of core functionality.

---

## 12. Future Expansion Architecture

Even in Phase 1, the data model and service interfaces are designed for:

```text
Organization (Phase 2+)
 ├── Store A
 │    ├── Kiosk
 │    ├── Kitchen
 │    ├── Cashier
 │    └── Online Orders
 ├── Store B
 └── Store C
```

Key design accommodations:
- `store_id` on all major tables
- Order source enum with future channels
- Abstract payment provider interface
- Role hierarchy that can expand (organization-level admin, etc.)

> **See [BRD.md](file:///c:/Users/Djhanggoo/Documents/Projects/FastFood%20System/fastFoodSystem/documentation/BRD.md) for the complete business requirements these technical decisions serve.**
