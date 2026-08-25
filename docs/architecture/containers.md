# Containers

**Target state** — see [target-architecture.md](target-architecture.md).  
**Today**: REST/MCP run in Python; graph-service and workers are not deployed.

| Container | Responsibility | Technology (target) |
|-----------|----------------|---------------------|
| **MemLore Core** | REST, MCP, authz boundary, governance transactions, outbox publish, retrieval orchestration, context compiler | Go 1.25+ (chi, pgx, sqlc) |
| **Graph Knowledge Service** | Stable MemLore graph API; Graphiti integration | Python 3.12 + FastAPI |
| **Graphiti** | Temporal graph knowledge engine | graphiti-core (in graph-service) |
| **Graph Store** | Entities, facts, embeddings, indexes | Neo4j 5.x |
| **Control Store** | Identity, scope, policy, provenance, audit, outbox | PostgreSQL 16+ |
| **Worker** | Outbox consumer, graph sync retries, background jobs | Go (`cmd/worker`) |
| **Telemetry** | Traces, metrics, logs | OpenTelemetry |

## Current deployment (development)

| Container | Status |
|-----------|--------|
| Python REST + MCP (`uv run memlore`) | **Running** — F001/F002 |
| MemLore Core (Go) | **Planned** — F101 skeleton |
| Graph Knowledge Service | Not built |
| Worker | Not built |
| PostgreSQL | Docker Compose |
| Neo4j | Docker Compose (unused by app) |
| Redis | Docker Compose (unused by app) |

Local development: `docker compose up -d` for Postgres, Neo4j, Redis.
