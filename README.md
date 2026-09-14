# TicketQueue

A Go backend for event ticket sales, with order processing built around a queue. Built as a hands-on study project to close three specific technical gaps — **N:N modeling**, **real idempotency**, and **real concurrency** — using a domain (ticket sales) that naturally forces all three to surface:

- **Real N:N:** users ↔ events ↔ tickets, with business rules (ticket limits per event, per user).
- **Real idempotency:** a purchase request can be resent by the client (timeout, network retry) without generating a duplicate charge.
- **Real concurrency:** limited ticket stock contested by multiple buyers at the same time — the classic "last ticket" race condition.

The project is being built incrementally, week by week, with technical decisions documented alongside the code. This README tracks that progress and will keep growing as the project evolves.

---

## Tech stack

- **Go 1.25**
- **Gin** — HTTP framework
- **PostgreSQL 15** (via Docker Compose)
- **database/sql** + **lib/pq** — database access with no ORM (a deliberate choice, see Architecture below)
- **golang-migrate** — versioned migrations in plain SQL
- **shopspring/decimal** — exact-precision monetary values (avoids `float64` rounding errors)
- **google/uuid** — identifiers

---

## Architecture

The project follows a variation of **ports & adapters (hexagonal architecture)**, organized by business domain:

```
internal/
  domain/<feature>/          → business logic + repository interface (the "port")
  infra/postgres/<feature>/  → concrete repository implementation on Postgres (the outbound "adapter")
  infra/http/<feature>/      → HTTP handlers (the inbound "adapter")
  infra/http/routes/         → route registration per domain
  config/                    → environment variable loading
cmd/main.go                  → composition root: wires config → connection → repository → service → handler → routes → server
```

Every domain (`users`, `events`, `orders`, `tickets`) follows the same dependency chain: **domain** declares the repository interface and the business logic (service), with no knowledge of HTTP or SQL; **infra/postgres** implements that interface against the database; **infra/http** translates HTTP requests into service calls and shapes the response. `main.go` is the only place that knows about every piece and wires them together manually (dependency injection with no framework).

**Why no ORM:** part of the point of this project is demonstrating real SQL and schema decisions (constraints, indexes, FK behavior) — using `database/sql` directly keeps those decisions explicit and visible instead of hidden behind an ORM.

---

## Data model

Five tables, all with `id UUID DEFAULT gen_random_uuid()` as primary key:

| Table | Role |
|---|---|
| `users` | users (unique `username`) |
| `events` | events (name, category, date, location) |
| `tickets` | ticket **types** per event (e.g. full price, half price), with `quantity_total`/`quantity_available` |
| `orders` | orders, with `status`, `idempotency_key` and `total_price` |
| `order_tickets` | N:N pivot table between `orders` and `tickets`, with `quantity` and `unit_price` (price snapshot at purchase time) |

Design decisions worth noting:

- **`price`/`total_price`/`unit_price` are `NUMERIC(10,2)`, never `FLOAT`** — binary floating point can't represent exact decimal values and accumulates rounding error on money.
- **`unit_price` is duplicated on `order_tickets`** instead of always being read from `tickets.price` — deliberate denormalization: if the ticket price changes later, the order's history must not change with it.
- **`tickets.quantity_total` and `quantity_available` coexist** — `available` is the mutable counter concurrency contends over; `total` preserves the original allocation and enables a sanity check (`total - available` should always match units sold).
- **`ON DELETE` behavior differs per relationship, on purpose:**
  - `events → tickets`: `CASCADE` — with no sales yet, deleting an event can clean up its ticket types too.
  - `order_tickets → tickets`: no cascade (default `RESTRICT`) — prevents deleting a ticket type that has already sold, which in turn also blocks deleting the parent event if it has sales (the `events` cascade hits this restriction).
  - `order_tickets → orders`: `CASCADE` — an order line item has no meaning on its own without the order.
- **`orders.idempotency_key` has `UNIQUE(user_id, idempotency_key)`**, not a global unique — the key is client-generated, so the guarantee is scoped per user.

---

## Running locally

Prerequisites: Go 1.25+, Docker, and the `golang-migrate` CLI installed with the Postgres driver:

```
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```
(make sure `$(go env GOPATH)/bin` is on your `PATH`)

1. Create a `.env` file at the project root:
   ```
   ENV=dev
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=postgres
   DB_NAME=go-ticket-queue
   SERVER_PORT=5001
   ```
2. Start Postgres: `docker compose up -d`
3. Apply migrations: `make migrate-up`
4. Run the server: `go run ./cmd`
5. Run the tests: `make test`

Useful `makefile` targets: `migrate-up`, `migrate-down`, `migrate-up-one`, `migrate-down-one`, `migrate-version`, `create-migration name=migration_name`, `test`.

---

## Available endpoints

| Method | Route | Description |
|---|---|---|
| `POST` | `/users` | Create a user |
| `GET` | `/users/:id` | Get user by ID |
| `GET` | `/users?username=` | Get user by username |
| `POST` | `/events` | Create an event |
| `GET` | `/events/:id` | Get event by ID |
| `POST` | `/orders` | Create an order — requires an `Idempotency-Key` header |

Still no business logic beyond the essentials (no auth, no `order_tickets` in the order-creation flow yet).

---

## Weekly progress

### Week 1 — Foundations and N:N modeling
- Modeled the 5 tables and versioned them with `golang-migrate`.
- Local Postgres via Docker Compose.
- Layered architecture (domain / infra-postgres / infra-http) established with `users` and `events`.
- Basic CRUD (create + fetch) for `users` and `events`, with no business rule beyond `username` uniqueness.
- "Not found" handling pattern: the repository translates `sql.ErrNoRows` into `(nil, nil)`; the handler decides the `404`.

### Week 2 — Real idempotency
- `POST /orders` reads the key from the `Idempotency-Key` header (not the request body — it's request metadata, not resource data).
- Idempotency strategy via **`INSERT ... ON CONFLICT (user_id, idempotency_key) DO NOTHING`**: a second attempt to create the same order neither errors nor duplicates a row — the `INSERT` simply does nothing, and the service fetches and returns the existing order instead.
- Deliberate decision **not** to do an "already exists?" check before creating: the actual protection lives entirely in the database constraint, which handles both retries and genuine concurrency the same way — checking first would only add a redundant query with no safety benefit.
- Integration test (`orders_service_test.go`) that fires 100 concurrent requests with the same idempotency key and proves, via `sync.WaitGroup` + a `COUNT(*)` check against the database, that only one order was created.

### Week 3 — Concurrency: reproducing the race condition
- Deliberately unsafe stock race simulation (`internal/domain/tickets/tickets_stock.go`): an in-memory `TicketStock` with a `Buy` method that reads, checks, and decrements quantity with **no protection at all**.
- Test (`tickets_stock_test.go`) fires 50 goroutines concurrently buying the same ticket (stock = 1).
- `go test -race` confirmed the data race (unsynchronized concurrent reads/writes on the `Quantity` and `BuysDone` fields), and the business assertion (`BuysDone > 1`) confirmed the real-world consequence: more than one "purchase" approved for a single ticket.
- Core takeaway: Go's `-race` detector only sees memory inside the Go process itself — a race happening purely through SQL against Postgres wouldn't be caught by it, which is why the simulation was built in memory before any database involvement.

### Coming up (weeks 4-7)
- **Week 4:** fix the race condition with `sync.Mutex`/`RWMutex`, with a before/after benchmark.
- **Week 5:** worker pool with channels to process orders asynchronously.
- **Week 6:** observability — structured logs (`slog`), basic metrics.
- **Week 7:** portfolio polish — deployment, documented decisions, walkthrough video.

---

## Tests

```
make test
```

Runs the whole test suite with environment variables loaded from `.env` (necessary because `go test` runs with each package's own directory as the working directory, not the project root). Integration tests (`orders`, `tickets`) expect a reachable Postgres via `.env`.

To reproduce week 3's concurrency analysis specifically:
```
go test -race ./internal/domain/tickets/... -run TestTicketsStock_RaceCondition -v -count=5
```
