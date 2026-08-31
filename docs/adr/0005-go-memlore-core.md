# ADR 0005: Go for MemLore Core

- **Status**: Accepted
- **Date**: 2026-08-25
- **Supersedes**: [ADR 0002](0002-python-fastapi-stack.md) for the **application control plane** (REST, MCP, workers, governance orchestration)

## Context

MemLore's governance plane (users, scopes, authority, verification, audit,
outbox, retrieval orchestration, MCP/REST) must be production-grade, strongly
typed, and operable at scale. The current Python FastAPI implementation
delivered the first vertical slices (F001 REST, F002 MCP) but was always a
bootstrap stack (ADR-0002).

Migration discovery (`docs/development/MIGRATION_DISCOVERY.md`) confirmed:

- Governance lore CRUD works in Python and must be preserved via strangler
  migration, not big-bang rewrite.
- Graphiti integration does not exist yet; it should land in a **thin Python
  graph knowledge service**, not in the Go core.
- PostgreSQL remains governance source of truth (ADR-0001).
- Domain MCP interface remains the agent contract (ADR-0003).

Go is a better long-term fit for MemLore Core: static typing, first-class
concurrency for context compilation, small deployable binaries, and ecosystem
support for sqlc/pgx, goose, chi, OpenTelemetry, and the Go MCP SDK.

## Decision

1. **MemLore Core** (control / governance plane) will be implemented in **Go
   1.25+** using:
   - `net/http` + **chi** for HTTP
   - **pgx** + **sqlc** for PostgreSQL (explicit SQL, no ORM-driven domain)
   - **goose** for migrations
   - **slog** for logging; **OpenTelemetry** for telemetry
   - Go MCP SDK for agent interface (future slices)
   - **testcontainers-go** for integration tests where appropriate

2. **Graph knowledge service** remains **Python 3.12 + FastAPI + Graphiti**,
   isolated behind stable MemLore-oriented HTTP contracts. Graphiti types MUST
   NOT leak into Go domain packages.

3. **Strangler migration**: Python REST/MCP remain operational until each Go
   slice passes characterization and contract tests. No deletion of Python
   behavior before Go replacement is verified.

4. **Repository layout** (emerges with F101, grows with features):

   ```text
   cmd/memlore/          # HTTP + MCP entry (future)
   cmd/worker/           # outbox worker entry (future)
   internal/domain/      # pure domain
   internal/application/
   internal/adapters/
   internal/infrastructure/
   db/queries/           # sqlc input
   migrations/           # goose SQL
   graph-service/        # Python (future)
   ```

5. **Python application** (`src/memlore/`) is **legacy** for governance features
   being migrated. It is **not** removed until the corresponding Go slice is
   verified. It may remain permanently for `graph-service/`.

## Alternatives Considered

- **Keep Python for core**: simpler short-term, but couples control plane to
  Graphiti's Python ecosystem and weakens concurrency/ops story for orchestration.
- **Rust core**: strong performance; higher team friction and slower iteration
  for initial migration slices.
- **Java/Kotlin gateway**: extra runtime and duplicated domain models.
- **Big-bang Go rewrite**: rejected — violates constitution and migration plan.

## Consequences

- ADR-0002 is **superseded for MemLore Core**; Python remains valid for the
  graph knowledge service.
- CI must run Go (`go test`, `go vet`) alongside existing Python checks during
  migration.
- Alembic migrations become legacy; goose owns forward schema for Go path.
  Initial goose migration ports `0001_lore_audit` schema verbatim.
- New governance features SHOULD land in Go once F101 skeleton is complete,
  unless explicitly deferred.
- Spec Kit feature **F101** (`specs/003-go-core-skeleton/`) defines the first
  Go slice: module skeleton, tooling, schema port — **no traffic switch**.

## Update (2026-08-31) — strangler complete

Go slices F104–F112, F111, F003, F106a, and F107 are the production path.
The legacy Python governance tree (`src/memlore/`, root pytest, Alembic) was
removed. Python remains **only** as `graph-service/`. CI no longer runs a
root FastAPI pytest job.

See `specs/017-retire-python-core/`.

## References

- `docs/development/MIGRATION_DISCOVERY.md`
- `docs/development/migration-inventory.md`
- `specs/003-go-core-skeleton/`
- ADR-0001 Dual-plane architecture
- ADR-0003 Domain MCP interface
