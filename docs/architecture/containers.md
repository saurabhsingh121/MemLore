# Containers

**Target state** — see [target-architecture.md](target-architecture.md).

| Container | Responsibility | Technology |
|-----------|----------------|------------|
| **MemLore Core** | REST, MCP, authz boundary, governance transactions, outbox publish, retrieval orchestration, context compiler | Go 1.25+ (chi, pgx, sqlc) |
| **Graph Knowledge Service** | Stable MemLore graph API; Graphiti integration | Python 3.12 + FastAPI (`graph-service/`) |
| **Graphiti** | Temporal graph knowledge engine | graphiti-core (in graph-service) |
| **Graph Store** | Entities, facts, embeddings, indexes | Neo4j 5.x |
| **Control Store** | Identity, scope, policy, provenance, audit, outbox | PostgreSQL 16+ |
| **Worker** | Outbox consumer, graph sync retries, background jobs | Go (`memlore worker`) |
| **Telemetry** | Traces, metrics, logs | OpenTelemetry |

## Current deployment (development)

| Container | Status |
|-----------|--------|
| MemLore Core (Go) | **Canonical** — `memlore serve` / `mcp` / `migrate` / `worker` |
| Graph Knowledge Service | Docker Compose `graph-service` |
| PostgreSQL | Docker Compose |
| Neo4j | Docker Compose |
| Redis | Docker Compose (optional) |

The legacy Python governance app (`src/memlore/`) has been **removed**.

Local development: `docker compose up -d postgres` (add `graph-service`/Neo4j
when using the knowledge plane).
