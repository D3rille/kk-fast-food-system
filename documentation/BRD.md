# Business Requirements Document (BRD)

**Project Name:** NextGen Kiosk-to-Kitchen Fast Food System

**Document Version:** 2.0

**Date:** June 9, 2026

**Author:** Systems Analyst

**Companion Document:** [PlanA.md](file:///c:/Users/Djhanggoo/Documents/Projects/FastFood%20System/fastFoodSystem/documentation/PlanA.md) — Technical Implementation Plan

---

## 1. Executive Summary

The objective of this project is to develop a robust, **local-first** base application for a self-service kiosk integrated directly with a Kitchen Display System (KDS). The system aims to minimize ordering friction, eliminate manual order-taking errors, and optimize kitchen throughput.

The system is architected around a **local-first, cloud-second** principle: all core operations (ordering, kitchen display, payment processing) must function over a local area network (LAN) without internet connectivity. Cloud synchronization is a secondary concern handled asynchronously when connectivity is available.

This document defines the **what** — business goals, functional requirements, data contracts, and acceptance criteria. The companion [PlanA.md](file:///c:/Users/Djhanggoo/Documents/Projects/FastFood%20System/fastFoodSystem/documentation/PlanA.md) defines the **how** — technology stack, architecture, and deployment strategy.

---

## 2. Project Scope

### 2.1 In-Scope (MVP — Phase 1)

* **Self-Service Kiosk UI:** Touch-optimized interface for browsing menus, customizing items, and queuing orders.
* **Kitchen Display System (KDS):** Real-time order fulfillment dashboard for kitchen staff.
* **Core Order State Machine:** Centralized business logic managing order lifecycles.
* **Payment Integration:** Abstracted payment provider system supporting cash, card terminals, and Philippine e-wallets (GCash, Maya).
* **Role-Based Administration:** JWT-based authentication with role-based access control for staff operations.
* **Menu & Administration Module:** CRUD operations for menu items, categories, modifiers, and real-time availability toggling.
* **Local-First Operation:** Full offline capability over LAN; cloud sync when internet is available.

### 2.2 Out-of-Scope (Phase 2+)

* Multi-tenant franchise/organization management (Phase 1 focuses on single-store architecture, but data models are multi-store ready).
* Advanced inventory predictive analytics.
* Integration with third-party delivery aggregators (Grab, Foodpanda, etc.).
* Customer accounts, loyalty programs, and OIDC/SSO authentication.
* Mobile app and online ordering channels (backend will be designed to support these, but UI is not built in Phase 1).

---

## 3. Business Architecture & Core Workflows

### 3.1 Order State Machine

All orders follow this lifecycle regardless of source channel:

```
[Draft] ──► [Pending Payment] ──► [Paid / Queued] ──► [In Preparation] ──► [Ready for Pickup] ──► [Completed]
                  │                                                                                    │
                  └──► [Cancelled]                                                    [Cancelled] ◄────┘
```

### 3.2 High-Level Order Flow

```
[Customer at Kiosk] ──(Places Order)──► [Order Service (Pending)] ──(Payment Success)──► [KDS Queue (Preparation)] ──(Staff Completes)──► [Customer Pickup]
```

### 3.3 Multi-Channel Order Source Design

Even though Phase 1 only implements the kiosk channel, the system must model order sources from day one to enable frictionless expansion:

```
Orders
 ├── kiosk      ← Phase 1
 ├── cashier    ← Phase 1
 ├── mobile     ← Phase 2+
 └── online     ← Phase 2+
```

The kitchen must be source-agnostic: orders from any channel appear identically on the KDS.

### 3.4 System Actor Matrix

| Actor              | Description                                               | Primary Goal                                                       |
|--------------------|-----------------------------------------------------------|--------------------------------------------------------------------|
| **Customer**       | End-user interacting with the physical kiosk.             | Place an accurate order quickly without line delays.                |
| **Kitchen Staff**  | Back-of-house team preparing food items.                  | View, prioritize, and fulfill orders based on FIFO.                |
| **Cashier**        | Staff member handling counter orders and cash payments.   | Enter orders manually and process cash/card payments.              |
| **Counter/Expeditor** | Staff member packing and handing over completed orders. | Verify order completion and hand over to the customer.             |
| **Store Manager**  | Administrative user managing day-to-day operations.       | Modify menu availability, manage staff roles, review reports.      |
| **Admin**          | System administrator with full access.                    | System configuration, user management, and store settings.         |

---

## 4. Functional Requirements

### 4.1 Kiosk Module (KM)

| ID       | Requirement                          | Description                                                                                                                                                           |
|----------|--------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| KM-001   | Idle Screen & Language Selection     | The kiosk must display an inviting idle screen. Tapping the screen prompts language selection.                                                                        |
| KM-002   | Menu Browsing                        | Display categories (e.g., Burgers, Drinks, Sides) and individual menu items with real-time pricing and images.                                                       |
| KM-003   | Item Modifier Engine                 | Customers can customize items based on predefined rules (e.g., "No Onions", "Upgrade to Large Fries"). Modifiers must support additive pricing and substitution logic.|
| KM-004   | Cart & Checkout Summary              | Customers can review, edit quantities, or delete items from their cart prior to finalizing.                                                                            |
| KM-005   | Ticket Generation                    | Upon payment confirmation, output a receipt with unique sequential order number (e.g., `#101`) and barcode/QR code for tracking.                                     |
| KM-006   | Combo/Meal Deal Support              | Support predefined meal combos with component selection (e.g., "Choose your drink", "Choose your side").                                                              |

### 4.2 Kitchen Display System Module (KDS)

| ID       | Requirement                          | Description                                                                                                                                                           |
|----------|--------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| KDS-001  | Real-Time Order Feed                 | New paid orders must instantly appear on the KDS screen via WebSocket push — no manual refresh required.                                                              |
| KDS-002  | Chronological Ticket Sorting         | Tickets sorted by timestamp (oldest first) with color-coded aging: **Green** (< 5 min), **Yellow** (5–10 min), **Red** (> 10 min).                                   |
| KDS-003  | Customization Highlights             | Modifiers, exclusions, and special instructions highlighted prominently (e.g., bolded red text for "NO PEANUTS") to prevent kitchen errors.                           |
| KDS-004  | Ticket State Transition              | Kitchen staff can touch/interact to transition tickets from *In Prep* → *Ready for Pickup*.                                                                          |
| KDS-005  | Source-Agnostic Display              | Orders from all channels (kiosk, cashier, future mobile/online) appear identically with a source indicator badge.                                                    |

### 4.3 Menu & Administration Module (AM)

| ID       | Requirement                          | Description                                                                                                                                                           |
|----------|--------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| AM-001   | Real-Time Item Kill-Switch           | Store managers can instantly mark an item as "Out of Stock", removing it from the kiosk view immediately.                                                            |
| AM-002   | Menu CRUD                            | Full management of items, base prices, categories, modifier groups, and modifier options.                                                                             |
| AM-003   | Product/Modifier Schema              | Flexible product/modifier schema supporting combos, size variants, add-ons, and substitutions from day one.                                                          |

### 4.4 Authentication & Authorization Module (AUTH)

| ID       | Requirement                          | Description                                                                                                                                                           |
|----------|--------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| AUTH-001 | JWT-Based Authentication             | All staff-facing endpoints require JWT authentication.                                                                                                                |
| AUTH-002 | Role-Based Access Control            | Enforce permissions based on roles: **Admin**, **Manager**, **Cashier**, **Kitchen**.                                                                                |
| AUTH-003 | Kiosk Public Access                  | Kiosk-facing menu browsing and order placement endpoints are publicly accessible (no customer login for Phase 1).                                                    |

### 4.5 Payment Module (PAY)

| ID       | Requirement                          | Description                                                                                                                                                           |
|----------|--------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| PAY-001  | Abstracted Payment Interface         | Payment processing must be provider-agnostic via an abstracted interface (create, verify, refund).                                                                   |
| PAY-002  | Supported Providers (Phase 1)        | Cash (at counter) and card terminal stub.                                                                                                                             |
| PAY-003  | Future Providers (Phase 2+)          | GCash, Maya, PayMongo, Stripe — pluggable without changing order logic.                                                                                              |

---

## 5. Non-Functional Requirements

### 5.1 Performance & Latency

| Metric                    | Target                                                                 |
|---------------------------|------------------------------------------------------------------------|
| KDS order display latency | ≤ 1.5 seconds from payment verification to KDS display                |
| UI touch responsiveness   | ≤ 100ms response time for touch targets                               |
| Backend startup time      | ≤ 3 seconds (critical for kiosk/restaurant environment)               |
| Memory footprint          | ≤ 256MB RAM for the backend service (suitable for Mini PC deployment) |

### 5.2 Reliability & Offline Operation

* **Local-First Architecture:** All core operations must function over LAN without internet. The kiosk, KDS, backend, and database all run on a local server.
* **Cloud Sync:** When internet connectivity is available, local data synchronizes to the cloud asynchronously. The system must handle reconnection gracefully.
* **No Split-Brain:** A local database/cache synchronization worker pattern must prevent data inconsistency between local and cloud states.

### 5.3 Maintainability & Observability

* **Structured Logging:** All order state transitions logged with structured fields (order ID, timestamp, actor, previous state, new state) for audit and performance metrics.
* **Health Endpoints:** `/healthz` (liveness) and `/readyz` (readiness) endpoints for monitoring.
* **Decoupled Architecture:** Event-driven communication (WebSockets) separating kiosk client from KDS UI. Internal services communicate via gRPC for typed contracts and performance.

### 5.4 Deployment Simplicity

* **Single-command deployment** via Docker Compose on a Mini PC or local server.
* No Kubernetes or complex orchestration for Phase 1.
* The system must be deployable by restaurant staff with minimal technical expertise.

---

## 6. Core Data Entities

The base application must support the following data relationships. The schema should be designed for single-store operation but with multi-store fields present for future expansion.

### 6.1 Entity Definitions

| Entity                | Key Fields                                                                                  |
|-----------------------|---------------------------------------------------------------------------------------------|
| **Store**             | ID, Name, Address, Timezone, IsActive                                                       |
| **User**              | ID, StoreID, Username, PasswordHash, Role (Enum), IsActive                                  |
| **Product**           | ID, Name, Description, BasePrice, CategoryID, ImageURL, IsAvailable                         |
| **Category**          | ID, Name, SortOrder, IsActive                                                               |
| **ModifierGroup**     | ID, Name, MinSelection, MaxSelection (e.g., "Choose 1 Drink Upgrade")                      |
| **ModifierOption**    | ID, ModifierGroupID, Name, ExtraPrice                                                       |
| **Order**             | ID, StoreID, OrderNumber, Source (Enum: kiosk/cashier/mobile/online), Timestamp, TotalAmount, OrderStatus (Enum), PaymentStatus (Enum) |
| **OrderItem**         | ID, OrderID, ProductID, Quantity, UnitPrice, CalculatedSubtotal                             |
| **OrderItemModifier** | ID, OrderItemID, ModifierOptionID                                                           |
| **Payment**           | ID, OrderID, Provider (Enum), Amount, Status (Enum), TransactionRef, Timestamp              |

### 6.2 Key Relationships

```
Store ──1:N──► User
Store ──1:N──► Order
Category ──1:N──► Product
Product ──N:M──► ModifierGroup (via ProductModifierGroup join)
ModifierGroup ──1:N──► ModifierOption
Order ──1:N──► OrderItem
OrderItem ──1:1──► Product
OrderItem ──1:N──► OrderItemModifier
OrderItemModifier ──1:1──► ModifierOption
Order ──1:N──► Payment
```

---

## 7. Risks & Assumptions

### Assumptions

* The physical kiosk hardware includes a reliable touch screen, a local thermal receipt printer, and operates on a stable local network.
* A single Mini PC or local server is available to host the backend, database, and serve the frontend.
* Phase 1 deployment targets a single store.

### Risks & Mitigations

| Risk                                                         | Impact | Mitigation                                                                                              |
|--------------------------------------------------------------|--------|---------------------------------------------------------------------------------------------------------|
| Network drops cause data inconsistency between local and cloud | High   | Local DB/cache sync worker pattern; queue local operations and replay to cloud on reconnection.          |
| Menu modifier schema is too rigid for real-world combos       | High   | Design a flexible product/modifier schema from day one; validate against real menu data early.           |
| Single point of failure on local server                      | Medium | Docker health checks with auto-restart; recommend UPS for the local server hardware.                    |

---

## 8. Acceptance Criteria Summary

| Module   | Criteria                                                                                     |
|----------|----------------------------------------------------------------------------------------------|
| Kiosk    | Customer can browse menu, customize items, add to cart, pay, and receive a printed receipt.   |
| KDS      | Kitchen staff see new orders in ≤ 1.5s, with aging indicators and modifier highlights.       |
| Admin    | Manager can toggle item availability and the change reflects on kiosks within 2 seconds.     |
| Auth     | Staff login with JWT; role-based access enforced across all protected endpoints.              |
| Payment  | Cash payment flow completes end-to-end; payment provider is swappable via interface.         |
| Offline  | Full ordering and kitchen display functionality continues when internet is disconnected.      |

---

## 9. Next Steps

1. Derive the Entity-Relationship Diagram (ERD) based on Section 6.
2. Define API contract definitions (OpenAPI/Swagger) for `/orders`, `/menu`, and `/auth` endpoints.
3. Establish the WebSocket hub architecture for KDS real-time events.
4. Implement the core order state machine and payment provider interface.
5. Build the kiosk and KDS frontends against the API contracts.

> **See [PlanA.md](file:///c:/Users/Djhanggoo/Documents/Projects/FastFood%20System/fastFoodSystem/documentation/PlanA.md) for the complete technical implementation plan.**