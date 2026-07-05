# Mock Data Seeder

The `mock-seed` command populates the database with a realistic set of development data so you can run the full system manually without setting up real store data.

## When to use

Run this once after `make migrate-up` whenever you want a clean set of test data. It is fully **idempotent** — running it multiple times is safe; all records are upserted by fixed UUID so duplicates are never created.

> **Warning:** The seeder targets the database pointed to by `DATABASE_URL`. Do not run against a production database.

## Usage

```bash
# Make sure migrations are applied first
make migrate-up

# Seed mock data
make mock-seed
```

The command prints a summary of what was created and the credentials you need:

```
Seeding mock data...
  ✓ store
  ✓ 4 users
  ✓ 5 categories
  ✓ 19 products

Done. Credentials:

  Store ID : aaaaaaaa-0000-0000-0000-000000000001

  Username   Password     Role
  ---------  -----------  --------
  admin      admin123     admin
  kiosk      kiosk123     cashier
  manager    manager123   manager
  kitchen    kitchen123   kitchen

Kiosk .env (web/kiosk/.env):
  VITE_STORE_ID=aaaaaaaa-0000-0000-0000-000000000001
  VITE_KIOSK_USERNAME=kiosk
  VITE_KIOSK_PASSWORD=kiosk123
```

## What gets seeded

| Entity | Count | Details |
|--------|-------|---------|
| Store | 1 | "NextGen Kitchen", SM Mall of Asia, Pasay City |
| Users | 4 | admin, kiosk (cashier), manager, kitchen |
| Categories | 5 | Burgers, Chicken, Sides & Snacks, Drinks, Desserts |
| Products | 19 | 3–4 items per category, prices in centavos |

## Configuring the kiosk frontend

After seeding, copy the kiosk env example and fill in the values printed by the seeder:

```bash
cp web/kiosk/.env.example web/kiosk/.env
```

Edit `web/kiosk/.env`:

```env
VITE_API_BASE_URL=http://localhost:8080
VITE_STORE_ID=aaaaaaaa-0000-0000-0000-000000000001
VITE_KIOSK_USERNAME=kiosk
VITE_KIOSK_PASSWORD=kiosk123
```

## Full manual test flow

```bash
# Terminal 1 — backend
make run

# Terminal 2 — kiosk frontend
cd web/kiosk && pnpm dev
```

1. Open `http://localhost:3000` — Welcome screen appears
2. Tap anywhere — Menu screen loads with Burgers, Chicken, Sides & Snacks, Drinks, Desserts
3. Tap **Add** on any product — cart badge increments in the header
4. Tap **Review Order** or the cart button — Cart screen shows items with +/− controls
5. Tap **Place Order** — Payment screen with total amount due
6. Tap **Cash at Counter** — order is created, checked out, and paid in sequence
7. Confirmation screen shows the order number

## Seeder implementation

Source: `cmd/seed/main.go`

The seeder uses the same `config.Load()` + `database.New()` path as the API server, so it respects all the same environment variables and `configs/.env` file. All records use fixed UUIDs so the seeder can be re-run safely via `ON CONFLICT (id) DO UPDATE`.
