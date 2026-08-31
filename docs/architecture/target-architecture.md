# Target Architecture

**Status**: Target state for MemLore production platform  
**ADR**: [ADR-0005 Go for MemLore Core](../adr/0005-go-memlore-core.md)  
**Current baseline**: [MIGRATION_DISCOVERY.md](../development/MIGRATION_DISCOVERY.md)

MemLore preserves engineering *why* — architectural decisions, rationale,
provenance, authority, verification, temporal knowledge, conflicts, and
evidence — so humans and coding agents retrieve trustworthy context with
receipts, not just similar text.

```text
                 Humans / Coding Agents
                          |
                     MCP / REST
                          |
                    MemLore Core
                        Go
                          |
        +-----------------+-----------------+
        |                                   |
        v                                   v
   PostgreSQL                         Graph Knowledge
 Control / Governance                    Service
       Plane                             Python
                                           |
                                        Graphiti
                                           |
                                         Neo4j
```

## MemLore Core (Go)

The control plane owns governance, orchestration, and agent-facing APIs.
It does **not** embed Graphiti or speak Neo4j directly.

| Concern | Owned by Go Core |
|---------|------------------|
| Identity & tenancy | users, teams, projects, repositories |
| Scoping | scope resolution, permissions |
| Lore governance | lore metadata, verification state, provenance |
| Authority | explainable evaluation at compile/explain (ephemeral; F003) |
| Evidence | references, strength metadata |
| Conflicts & supersession | detection, policy, governance records |
| Audit | append-only history |
| Retrieval orchestration | parallel fetch, ranking, context compilation |
| Agent interface | MCP (`memlore.*`) |
| Integrations | REST, CLI |
| Async sync | transactional outbox, background workers |
| Graph coordination | calls graph-service over HTTP/gRPC port |

**Stack (baseline)**: Go 1.25+, chi, pgx, sqlc, goose, slog, OpenTelemetry,
Go MCP SDK, testcontainers-go.

**Dependency rule**:

```text
Adapters → Application → Domain
                ↑
         Ports ← Infrastructure
```

Domain packages MUST NOT import HTTP, MCP, PostgreSQL, Neo4j, Graphiti, or
OpenTelemetry.

## Graph Knowledge Service (Python)

A deliberately thin service isolating Graphiti behind MemLore-oriented contracts.
Graphiti is an implementation detail — not part of the domain model.

| Concern | Owned by graph-service |
|---------|------------------------|
| Episode ingestion | normalize and ingest episodes |
| Entity extraction | Graphiti pipeline |
| Facts & relationships | temporal graph writes |
| Semantic retrieval | embedding + graph search |
| Temporal invalidation | fact lifecycle in graph store |

**Conceptual API** (contracts defined in specs + tests):

| Method | Purpose |
|--------|---------|
| `POST /episodes` | Ingest an episode |
| `POST /search` | Semantic / graph search |
| `GET /facts/{id}` | Retrieve a fact |
| `POST /facts/{id}/invalidate` | Mark fact invalid |
| `POST /facts/{id}/supersede` | Supersede with successor |

**Stack**: Python 3.12+, FastAPI, Pydantic, Graphiti, Neo4j, OpenTelemetry,
pytest.

Go communicates via a `KnowledgeGraph` port — never imports Graphiti types.

## Data ownership

| Store | Source of truth for |
|-------|---------------------|
| **PostgreSQL** | users, teams, projects, repos, scopes, lore metadata, authority factors, verification, provenance, evidence refs, permissions, audit, ingestion status, outbox events |
| **Neo4j (via Graphiti)** | entities, facts, relationships, episodes, embeddings, temporal graph retrieval |

PostgreSQL is authoritative for governance decisions even when graph search
returns candidates.

## Synchronization (no distributed transactions)

```text
API request
   |
PostgreSQL transaction
   |
   +--- save Lore metadata
   |
   +--- save outbox event
   |
COMMIT
   |
worker (Go)
   |
graph-service (Python)
   |
Graphiti / Neo4j
```

Workers MUST be idempotent and retry-safe. Never require Postgres + Neo4j to
commit atomically.

## Context compilation pipeline

Primary agent operation: **`memlore.get_for_task`**

```text
request (repo, branch, task, query, files, token budget)
   ↓
scope resolution
   ↓
parallel retrieval ──┬── PostgreSQL (authority metadata)
                     ├── graph-service search
                     └── repo / task metadata
   ↓
temporal filtering (superseded / invalidated)
   ↓
conflict detection
   ↓
authority evaluation + ranking / dedup
   ↓
token budgeting
   ↓
ContextPacket
```

v1 (F003) evaluates authority **after** the temporal filter on the compile
path so stale lore is not scored by default. `memlore.explain` still evaluates
the fetched entry, including history.

## MCP surface (target)

Agents interact with MemLore, not Graphiti directly.

| Tool | Purpose |
|------|---------|
| `memlore.get_for_task` | Compiled context for a task (preferred) |
| `memlore.repo_profile` | Compact repository intelligence profile |
| `memlore.search` | Search with governance filters |
| `memlore.remember` | Store lore |
| `memlore.get` | Fetch by id |
| `memlore.verify` | Mark verified |
| `memlore.explain` | Provenance + authority reasoning |
| `memlore.supersede` | Supersede knowledge |
| `memlore.invalidate` | Invalidate knowledge |

## Security boundaries

- Treat all stored context as potentially untrusted (prompt injection, malicious memories).
- Authority ≠ authorization.
- Enforce tenant and repository isolation at the Go core.
- Never expose Graphiti/Neo4j internals via MCP.

## Current vs target

| Layer | Today (main) | Target |
|-------|--------------|--------|
| Core runtime | **Go** MemLore Core | Go (unchanged) |
| REST / MCP | Go `memlore serve` / `memlore mcp` | Go adapters |
| Governance DB | PostgreSQL ✓ | PostgreSQL ✓ |
| Graph service | Thin Python `graph-service/` | Same |
| Graphiti / Neo4j | graph-service + Docker Compose | Same |
| Outbox / workers | Go `memlore worker` | Go worker |
| Context compiler | Go application layer | Same |

Strangler migration is complete (F113). Membership-scoped authz is complete
(F114). F020 repository intelligence profile is on the Go core. Remaining
product work is the engineering-intelligence flywheel (F021+): richer
`get_for_task`, git/PR/ADR ingest, suggested-lore review (CLI + REST),
first-class decisions, `memlore why`, architecture drift, and GitHub PR
checks. See [FEATURE_DEVELOPMENT.md](../development/FEATURE_DEVELOPMENT.md).

## Related documents

- [Overview](overview.md) — planes and adapters (includes current delivery notes)
- [Containers](containers.md) — deployable units
- [Authority model](authority-model.md)
- [Product feature roadmap](../development/FEATURE_DEVELOPMENT.md)
- [ADR-0001 Dual-plane](../adr/0001-dual-plane-architecture.md)
- [ADR-0003 Domain MCP](../adr/0003-mcp-domain-interface.md)
- [ADR-0005 Go core](../adr/0005-go-memlore-core.md)
