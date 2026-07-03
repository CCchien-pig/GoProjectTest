# UDM Platform Walkthrough (Weeks 1 & 2 Completed)

We have successfully implemented and compiled the entire code base for **Week 1 (PostgreSQL CRUD & Base Infrastructure)** and **Week 2 (ScyllaDB Time-Series Telemetry & Alert Events)** using a Test-Driven Development (TDD) workflow.

---

## 🛠️ Key Implemented Components

### 1. Base Infrastructure & Configurations (Week 1 — Day 1 & 2)

- **Config & Environments**: Initialized dynamic loading of variables from `.env.dev` via `godotenv`.
- **API Formatting**: Standardized JSON REST response wrappers (`OK`, `OKWithPagination`, `BadRequest`, `NotFound`, `InternalError`, `ServiceUnavailable`).
- **Trace ID**: Request ID tracing middleware injecting `X-Request-ID` headers to all API endpoints.
- **SQL Migrations**: Schema migrations generated for `users`, `devices`, and `alert_rules` with GIN `pg_trgm` indexes.

### 2. GORM & PostgreSQL CRUD (Week 1 — Day 3 to 5)

- **Users**: Create, Update, Read (returning count of owned devices), and Soft Delete (`is_active = false`) via GORM.
- **Devices**: CRUD with **Cursor-based pagination** (using `(created_at, id)` tuples) and pg_trgm fuzzy matching search.
- **Alert Rules**: CRUD supporting rule parameters validation.
- **Timestamp Hook**: Integrated GORM's `BeforeUpdate` hook to automatically set `updated_at` on updates.

### 3. ScyllaDB & Time-Series Data (Week 2 — Day 6 to 9)

- **ScyllaDB Client**: Initialized connection with automatic keyspace setup and tables creation (`telemetry` with 90-day TTL, `alert_events` with 365-day TTL).
- **Telemetry Ingestion**: Ingestion endpoint processing batch writes. Includes **Alert Trigger logic**: when an ingested telemetry metric crosses a threshold defined in PostgreSQL rules, an `alert_event` is automatically registered to ScyllaDB.
- **Queries & Partition Splits**: Implemented range queries that split the query into day-based partition keys, perform concurrent reads, and merge results.
- **Soft Deletion Check**: Resolves whether the queried device was deleted in PostgreSQL, returning time-series data with a flag `is_deleted: true`.
- **Embed Latest Telemetry**: Expanded device details retrieval (`GET /api/v1/devices/:id`) to retrieve and embed latest telemetry parameters directly from ScyllaDB.
- **Alert Event ACK**: Handled acknowledgement (`acknowledged = true`) targeting clustered partition indexes.

### 4. Integration Wiring & Decoupling (Week 2 — Day 10)

- **Routes setup**: Standardized all endpoints in [routes.go](file:///c:/Projects/CC/Go/GoProjectTest/internal/routes/routes.go).
- **Main Assembly**: Structured DI layer and graceful shutdowns in [main.go](file:///c:/Projects/CC/Go/GoProjectTest/cmd/api/main.go).
- **Offline / Degraded Mode Safety**: Configured services to catch connection failures on ScyllaDB/KeyDB and switch to degraded mode cleanly (HTTP 503 Service Unavailable for ingestions/queries) while keeping PostgreSQL device details functioning.

---

## 🧪 Verification & Validation

All tests in the project run and pass successfully!

```bash
go test ./... -v
```

```
?    GoProject/udm/cmd/api [no test files]
=== RUN   TestLoad
--- PASS: TestLoad (0.00s)
PASS
ok   GoProject/udm/internal/config 0.739s
=== RUN   TestAlertRuleHandler_Create
--- PASS: TestAlertRuleHandler_Create (0.19s)
=== RUN   TestAlertRuleHandler_FindByDeviceID
--- PASS: TestAlertRuleHandler_FindByDeviceID (0.00s)
=== RUN   TestDeviceHandler_Create
--- PASS: TestDeviceHandler_Create (0.00s)
...
=== RUN   TestTelemetryHandler_Query
--- PASS: TestTelemetryHandler_Query (0.00s)
PASS
ok   GoProject/udm/internal/handler 7.288s
?    GoProject/udm/internal/keydb [no test files]
=== RUN   TestTraceID
--- PASS: TestTraceID (0.00s)
PASS
ok   GoProject/udm/internal/middleware 0.669s
...
=== RUN   TestTelemetryService_BatchInsert_AlertTrigger
--- PASS: TestTelemetryService_BatchInsert_AlertTrigger (0.17s)
PASS
ok   GoProject/udm/internal/service 1.688s
=== RUN   TestResponseHelpers
...
--- PASS: TestResponseHelpers (0.00s)
PASS
ok   GoProject/udm/pkg/response 0.686s
```
