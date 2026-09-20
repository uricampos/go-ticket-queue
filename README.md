# TicketQueue

A Go backend for event ticket sales, with order processing built around a queue. Built as a hands-on study project to close three specific technical gaps — **N:N modeling**, **real idempotency**, and **real concurrency** — using a domain (ticket sales) that naturally forces all three to surface:

- **Real N:N:** users ↔ events ↔ tickets, with business rules (ticket limits per event, per user).
- **Real idempotency:** a purchase request can be resent by the client (timeout, network retry) without generating a duplicate charge.
- **Real concurrency:** limited ticket stock contested by multiple buyers at the same time — the classic "last ticket" race condition.

The project was built incrementally, week by week, with technical decisions documented alongside the code — the full log is below, from initial N:N modeling through to a live deployment.

**Live demo:** [go-ticket-queue.uricampos.dev](https://go-ticket-queue.uricampos.dev) · **API reference:** [go-ticket-queue.uricampos.dev/docs](https://go-ticket-queue.uricampos.dev/docs)

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
| `GET` | `/orders/processed` | Count of orders processed so far |
| `GET` | `/orders/queue-size` | Current number of order jobs waiting for a free worker |
| `GET` | `/docs` | Interactive API reference (Scalar), generated from `openapi.yaml` |
| `GET` | `/openapi.yaml` | The raw OpenAPI 3.0 specification |

`/users`, `/events` and `/orders` are all rate-limited per client IP (see week 7 below); `/docs` and `/openapi.yaml` are not. Still no business logic beyond the essentials (no auth, no `order_tickets` in the order-creation flow yet).

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

### Week 4 — Concurrency: fixing the race condition with a mutex
- Added a `sync.Mutex` to `TicketStock` (embedded by value as an unexported field, `sm`) — the zero-value mutex needs no separate construction, and keeping it unexported stops external code from locking/unlocking it directly, bypassing `Buy`.
- The lock wraps the entire critical section of `Buy` — from reading `Quantity` through deciding `canBuyTicket` to writing the decrement — so the read-check-write sequence is now atomic from every other goroutine's point of view.
- Reran the race test from week 3 five times in a row with `-race`; zero races detected:
  ```
  go test -race ./internal/domain/tickets/... -run TestTicketsStock_RaceCondition -v -count=5
  ```
- **Why the lock fixes it:** before, two goroutines could interleave so that both read `Quantity=1` before either wrote back, both concluded "I can buy", and both decremented — a classic lost-update race. With the mutex, one goroutine fully completes its read-check-write (and releases the lock) before another can even start evaluating the condition, so every read always reflects the latest write.
- Wrote `BenchmarkTicketStock_Buy` using `b.RunParallel` (Go's tool for measuring throughput under real concurrent load, as opposed to a single-goroutine benchmark) and measured it under four conditions — with/without the mutex, and with/without the artificial `time.Sleep(1ms)` that week 3 used to widen the race window:
  ```
  go test -bench=. -run=^$ ./internal/domain/tickets/...
  ```

  Environment: `goos: linux`, `goarch: amd64`, `cpu: 13th Gen Intel(R) Core(TM) i7-13620H` (16 logical cores — reflected in the `-16` suffix on the benchmark name, i.e. `GOMAXPROCS=16`).

  | | With mutex | Without mutex |
  |---|---|---|
  | **With artificial 1ms sleep** | 1,160,241 ns/op | 72,855 ns/op |
  | **Without sleep** | 404.1 ns/op | 30.37 ns/op |

  Only same-row comparisons are apples-to-apples (a single variable — the mutex — toggles; the sleep condition is held constant):
  - **With the sleep:** removing the mutex is ~16x faster (72,855 vs 1,160,241 ns/op). Without the lock, all goroutines sleep concurrently across cores; with it, they sleep one at a time, serializing work that would otherwise run in parallel.
  - **Without the sleep:** the mutex still costs ~13x more (404.1 vs 30.37 ns/op) even though the protected work is trivial (comparing and decrementing two ints). That gap is the real cost of **lock contention** — 16 goroutines competing for the same mutex — not the cost of the lock primitive itself (an uncontended `Lock`/`Unlock` pair alone costs tens of nanoseconds, not hundreds).
  - Takeaway: a mutex's cost scales with how many goroutines are fighting over it, not just with how much work the critical section does — protecting even trivial operations under heavy concurrent access has a real, measurable price.

### Week 5 — Concurrency: the same problem solved with a channel-based worker
- Implemented a second, alternative fix for the exact same race from weeks 3-4, this time using Go's other concurrency tool: channels, instead of a mutex.
- `buyRequest` (`tickets_stock.go`) is a small message type carrying the purchase quantity and its own dedicated response channel (`result chan bool`) — the caller uses that private channel to get its individual answer back.
- `StartWorker` launches a single dedicated goroutine that owns the `TicketStock` exclusively: it loops over an incoming `chan buyRequest` (`for req := range requests`), processing one request at a time. No mutex anywhere inside it — there's nothing to protect, because no other goroutine ever touches `Quantity`/`BuysDone` directly.
- Callers (`TestTicketStock_RaceConditionChan`) never call `Buy` directly anymore; each one builds a `buyRequest` with its own response channel, sends it on the shared channel, and blocks on its own channel until the worker replies.
- Confirmed race-free the same way as week 4 — five consecutive `-race` runs, zero races detected, plus the same business assertion (`BuysDone > 1`) passing:
  ```
  go test -race -run TestTicketStock_RaceConditionChan -v ./internal/domain/tickets/... -count=5
  ```
- **Mutex vs. channel — same bug, two different fixes:**
  - A **mutex** lets many goroutines keep direct access to the shared state, and a lock decides who gets to touch it at any instant — the safety comes from the lock discipline everyone must follow.
  - A **channel worker** removes shared access altogether: exactly one goroutine owns the state permanently, and every other goroutine only ever sends it a message describing what it wants done. This is the Go proverb in practice — *"don't communicate by sharing memory; share memory by communicating."*
  - **When each fits better:** a mutex is simpler and cheaper for short, direct, synchronous access to shared state (like this one). A channel-based worker earns its complexity when you also want queuing, backpressure, or asynchronous processing decoupled from the caller — which is closer to what a real order-processing pipeline (the "Queue" in TicketQueue) needs. Multiple *independent* workers (a true worker pool) make sense for parallelizing unrelated tasks; for protecting one shared resource specifically, a single owning worker is the correct scale — adding more workers here would just reintroduce the original race.

### Week 5 (continued) — wiring the worker pool into real order processing
- Before integrating, revisited whether mutex/channel (the tools from weeks 4-5) were even the right fix for the *real* stock problem. They aren't: a mutex or a channel-owned worker only synchronizes goroutines **inside one process**. Once the shared state is a row in Postgres, the correct guarantee — one that also holds if the API ever runs as multiple instances behind a load balancer — is a single atomic, conditional `UPDATE`:
  ```sql
  UPDATE tickets
  SET quantity_available = quantity_available - $1
  WHERE id = $2 AND quantity_available >= $1
  RETURNING quantity_available
  ```
  This is the same lesson as week 2's idempotency design: the database's own concurrency control is the actual source of truth, not an in-process lock. Wiring this specific update into the purchase flow (via `order_tickets`) is still a TODO — what follows integrates the worker *pool pattern* itself into real order creation, which solves a different, still-relevant problem: bounding concurrent processing load, not protecting a shared counter.
- `OrderService` now owns a `jobs chan orderJob` and starts a fixed pool of worker goroutines (`StartWorkers`) inside its own constructor, `NewOrderService`. Neither the HTTP handler nor `main.go` changed a single line — `CreateOrder` kept its exact signature; internally, it now builds an `orderJob` with its own dedicated result channel, sends it on `s.jobs`, and blocks until a worker replies. The queue is entirely an implementation detail of the service.
- Unlike the single-owner worker from the `TicketStock` exercise, this pool intentionally runs **multiple** workers — order-creation jobs are independent of each other (unlike one shared stock counter), so parallelizing across a fixed pool is the correct shape here, capping how many orders are processed concurrently without limiting how many can be *accepted* at once.
- Load-tested the real `POST /orders` endpoint with a small throwaway script, `cmd/loadtest`: fires 2,000 goroutines at the running server, each with its own HTTP request, tracking status codes and elapsed time into shared slices (same index-per-goroutine pattern as the earlier concurrency tests). The idempotency key is a `main`-level variable passed into every call, making it trivial to switch between "2,000 requests, same key" and "2,000 requests, 2,000 unique keys":
  ```
  go run ./cmd/loadtest
  ```
  An initial attempt compared the repeated-key case (via `hey`) against the unique-key case (via this script) and produced a confusing, direction-flipping result — a sign the two *tools* weren't applying concurrency the same way (`hey` paces requests with `-c`; the script fires all goroutines at once), not a real effect of the key. Rerunning **both** conditions through the same script removed that confound:

  | | Repeated key | Unique keys |
  |---|---|---|
  | **5 workers** | 268.6 ms | 835 ms |
  | **50 workers** | 246.5 ms | 368.6 ms |

  All 8,000 requests across every run returned `201` with zero errors.

  - **Repeated key barely changes with more workers** (268.6→246.5 ms, ~8%): all 2,000 requests contend for the *same* row, and Postgres serializes access to that row internally via its own row-level lock — no matter how many application-level workers are free, they still queue at the database for that one row. The bottleneck isn't the pool size here, it's the database's own concurrency control on a single hot row.
  - **Unique keys improve substantially with more workers** (835→368.6 ms, ~56%): with no artificial contention on a shared row, the database can genuinely write many different rows in parallel, so the worker pool's size becomes the real limiting factor.
  - **Takeaway:** a worker pool's size matters most when the underlying work is actually parallelizable. When work is serialized by contention on a shared resource regardless of application-level concurrency, throwing more workers at it barely helps — the bottleneck has already moved somewhere else (here, a single Postgres row lock).

### Week 6 — Observability
- Configured a single global `slog` handler once, in `main.go` (`slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))`), so every log line across the app comes out as structured JSON instead of loose `fmt.Println` text.
- Instrumented the full lifecycle of an order so it can be reconstructed from logs alone: `order job enqueued` (right after the job is handed to the shared channel) → `order being processed by worker` (worker id + which order) → either `order created` or `order already existed, idempotency conflict resolved` (worker distinguishes the two outcomes explicitly, instead of leaving it to be inferred from the absence of a log line) → `order processed` / `order processing failed` (final outcome back in `CreateOrder`, with the error attached on failure). A single request produces a readable, ordered trace like:
  ```json
  {"msg":"order job enqueued","user_id":"...","idempotency_key":"..."}
  {"msg":"order being processed by worker","worker_id":1,"user_id":"...","idempotency_key":"..."}
  {"msg":"order already existed, idempotency conflict resolved","user_id":"...","idempotency_key":"...","order_id":"...","status":"pending"}
  {"msg":"order processed","user_id":"...","idempotency_key":"...","order_id":"...","status":"pending"}
  ```
- Added two metrics endpoints: `GET /orders/processed` (a `COUNT` query against the database) and `GET /orders/queue-size` (an in-process `atomic.Int64` — Go's third concurrency primitive used in this project, after mutex and channel — incremented right before a job enters `s.jobs` and decremented the instant a worker picks it up). The queue-size counter exists because an **unbuffered channel has no meaningful length** (`len()` on it is always `0` — a send only completes when a receive is happening at that exact instant, so nothing ever sits "inside" it to count); tracking queue depth for an unbuffered channel requires a counter kept alongside it, not the channel itself.
- Reviewed the codebase for the classic N+1 query pattern (fetch a list, then query again per item) and concluded it structurally cannot occur here yet: every endpoint fetches exactly one row (by primary key or a unique constraint) or a single aggregate — there is no list-returning endpoint anywhere, so there is no "N" to multiply. Also confirmed no `SELECT *` exists anywhere in the codebase. Noted as a forward-looking risk, not a current one: an eventual "list a user's orders with their tickets" endpoint would need a `JOIN` or a single `WHERE ... IN (...)` query, not a per-row loop.

Two bugs worth naming because they were easy to miss and had unusually severe failure modes:
  - `OrderService.GetOrdersProcessedCount` originally called `s.GetOrdersProcessedCount` — itself, instead of `s.repo.GetOrdersProcessedCount` — an infinite recursion that crashed the process with `fatal error: stack overflow` on the very first call.
  - Passing `OrderService` **by value** into `NewOrderHandler` (a pattern already in use for `UserHandler`/`EventHandler`) broke the moment `OrderService` gained its `atomic.Int64` field: `go vet` caught it as `NewOrderHandler passes lock by value` — atomic types embed a `noCopy` guard specifically to catch this, since copying one after use silently breaks its atomicity. Fixed by injecting `*OrderService` instead of a value everywhere.

### Week 7 — API documentation, rate limiting, and a real deployment

**API documentation (`openapi.yaml` + Scalar).** Wrote a full OpenAPI 3.0 specification for every endpoint — request bodies, headers (`Idempotency-Key`), status codes per outcome, and reusable `components/schemas` (`User`, `Event`, `Order`) referenced via `$ref` instead of repeated inline definitions. Rendered it with [Scalar](https://github.com/scalar/scalar), which turns a spec file into a full interactive reference (sidebar navigation, "try it" requests, examples) with a single embed script — no separate frontend project, no build step:
```html
<script id="api-reference" data-url="/openapi.yaml"></script>
<script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
```
Both `docs.html` and `openapi.yaml` are served directly by the Go binary via `router.StaticFile`, registered outside the rate-limited route groups.

**Rate limiting.** A per-client-IP token-bucket limiter (`golang.org/x/time/rate`), applied as Gin middleware on the `/users`, `/events` and `/orders` route groups. The state — one `*rate.Limiter` per IP — lives in a `map[string]*rate.Limiter` guarded by a `sync.Mutex`, Go's *fourth* concurrency tool used in this project (after mutex-on-a-struct, channel-based worker, and `atomic.Int64`), applied here to protect a shared map instead of a counter or a domain object. Three bugs surfaced while wiring it up, all textbook nil-value mistakes:
  - `var ipMutex *sync.Mutex` and `var ipList map[string]*rate.Limiter` — both declared with `var` instead of initialized (`&sync.Mutex{}`, `make(...)`) — a nil pointer panics on `.Lock()`, a nil map panics on write.
  - The lookup helper created a new limiter for an unseen IP, stored it in the map, but returned the *original* (nil) variable instead of the new one — every first-time visitor got a nil `*rate.Limiter` back.
  - The classic missing `return` after `ctx.Abort()`: without it, a rejected request still fell through to `ctx.Next()` and got processed anyway.

  Tuned the limits for the actual audience this API has — not production traffic, but a recruiter or interviewer clicking around the Scalar docs. An aggressive limit (the initial guess was 0.5 req/s) would visibly throttle someone just exploring the demo; a generous one (2 req/s, burst 15) still stops a scripted flood cold while being invisible to a human clicking buttons.

**Deployment.** Shipped to an existing VPS that already runs other projects behind a shared `nginx`, as a plain Go binary managed by `systemd` (no Docker for the app itself):
  - Cross-compiled locally (`GOOS=linux GOARCH=amd64 go build`) and copied the binary over — no Go toolchain needed on the server.
  - A **dedicated** `docker-compose` Postgres instance for TicketQueue (`ticketqueue-db`), isolated from the VPS's other projects' shared database — deliberately *not* reusing the existing shared Postgres that other apps on the box already depend on, so this project's data and uptime don't couple to theirs. Bound to `127.0.0.1` only, never exposed to the internet, since the Go binary reaches it over localhost.
  - Migrations applied from the local machine through an SSH tunnel to the VPS's Postgres port, rather than installing `migrate` or exposing the database port publicly.
  - A `systemd` unit (`go-ticket-queue.service`) runs the binary, restarts it on failure, and starts it on boot.
  - `nginx` reverse-proxies `go-ticket-queue.uricampos.dev` to the app's local port, with `certbot` issuing and auto-renewing the TLS certificate — added as a new, independent site alongside the VPS's existing ones, without touching their configuration.
  - Caught a real port collision before it caused an outage: the app's local `.env` defaults to port `5001`, which was already bound to an unrelated service on that VPS. Production uses a different port, configured only in the server's own `.env` — no change needed locally.

### Closing the loop

The three gaps this project set out to close are each demonstrably closed, not just described: **N:N modeling** in the five-table schema and its deliberate FK/denormalization choices (weeks 1-2); **real idempotency**, proven under actual concurrent HTTP load, not just unit-tested (week 2, and again under the load tests in week 5); and **real concurrency**, solved two different ways (mutex, then channel-based worker) on a reproduced-on-purpose race condition, benchmarked, and finally applied to real order processing bounded by a worker pool (weeks 3-5). Observability (week 6) and a public, documented, rate-limited deployment (week 7) turn the whole thing from a local exercise into something that can be handed to someone else to poke at.

Natural next steps exist but are deliberately out of scope here — adding gRPC or a real message broker (Kafka/RabbitMQ) on top of this same project, as the next things that show up often in Go backend roles in Europe. This project's job was the fundamentals underneath those, not the fundamentals plus everything else.

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

To reproduce week 4's throughput benchmark:
```
go test -bench=. -run=^$ ./internal/domain/tickets/...
```

To reproduce week 5's channel-based worker analysis:
```
go test -race -run TestTicketStock_RaceConditionChan -v ./internal/domain/tickets/... -count=5
```
