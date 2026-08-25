# Feature Specification: Graph Knowledge Service (F106)

**Feature Branch**: `009-graph-service`  
**Created**: 2026-08-25  
**Status**: Ready  
**Depends on**: F106a (Go governance hardening)

## Goal

Extract a thin Python `graph-service/` that isolates Graphiti and Neo4j behind
MemLore-oriented HTTP contracts. Go Core defines a `KnowledgeGraph` port and
HTTP client only — no Graphiti types in Go.

## User Scenarios & Testing

### User Story 1 — Service health (Priority: P1)

Operators and Go Core verify graph-service and Neo4j are reachable before
ingest or search.

**Independent Test**: `GET /health` returns `ok` when Neo4j is up.

**Acceptance Scenarios**:

1. **Given** Neo4j healthy, **When** `GET /health`, **Then** status is `ok`.
2. **Given** Neo4j down, **When** `GET /health`, **Then** service reports degraded Neo4j status.

---

### User Story 2 — Episode ingest (Priority: P1)

MemLore (future outbox worker) ingests lore statements into the knowledge graph
with scope and provenance metadata.

**Independent Test**: `POST /episodes` persists via Graphiti; integration test
round-trip proves retrieval.

**Acceptance Scenarios**:

1. **Given** valid episode payload, **When** `POST /episodes`, **Then** returns `episode_id`.
2. **Given** missing statement or scope, **When** `POST /episodes`, **Then** returns validation error.
3. **Given** ingested episode, **When** `POST /search` with related query, **Then** MemLore-shaped results include the fact.

---

### User Story 3 — Semantic search (Priority: P2)

Retrieval orchestration (F108) queries graph-service with MemLore filters, not
raw Graphiti shapes.

**Independent Test**: `POST /search` returns `{ results: [...] }` with `id`,
`statement`, `score`, optional `scope`, no Graphiti field names.

**Acceptance Scenarios**:

1. **Given** indexed episodes, **When** `POST /search` with query, **Then** non-empty ranked results.
2. **Given** scope filter, **When** `POST /search`, **Then** results respect scope group.

---

### User Story 4 — Go boundary (Priority: P2)

Go Core can call graph-service via `KnowledgeGraph` port without importing
Graphiti.

**Independent Test**: Go contract test (integration tag) hits health + episode
against running service.

**Acceptance Scenarios**:

1. **Given** running graph-service, **When** Go client `Health`, **Then** no error.
2. **Given** running graph-service, **When** Go client `IngestEpisode`, **Then** returns episode id.

---

### Edge Cases

- Neo4j unavailable: health degraded; ingest/search return 503 with clear error envelope.
- OpenAI API key missing: Graphiti ingest fails; integration tests skip.
- Empty search results: return `{ "results": [] }`, not error.
- Invalid fact id on `GET /facts/{id}`: return 404.

## Requirements

### Functional Requirements

- **FR-001**: `graph-service/` FastAPI app with uv, ruff, mypy --strict, pytest.
- **FR-002**: `GET /health` reports service and Neo4j connectivity.
- **FR-003**: `POST /episodes` accepts statement, scope, metadata, provenance_refs.
- **FR-004**: `POST /search` accepts query and optional scope filter; returns MemLore results.
- **FR-005**: `GET /facts/{id}` returns fact or 404 (stub acceptable for v1).
- **FR-006**: Graphiti types confined to `graph_service/adapters/graphiti/`.
- **FR-007**: OpenAPI in `graph-service/openapi.yaml`; docs in `docs/api/graph-service.md`.
- **FR-008**: Go `KnowledgeGraph` port + HTTP client in `internal/infrastructure/graphclient/`.
- **FR-009**: Contract tests: Python API shapes; Go client integration test.
- **FR-010**: Docker Compose wires graph-service + Neo4j; CI job for graph-service quality.

### Key Entities

- **Episode**: Ingest unit — statement, scope, metadata, provenance refs, reference time.
- **GraphFact**: Retrieval unit — id, statement, score, scope, provenance refs.
- **Scope**: kind + key (aligned with governance plane scopes).

## Success Criteria

- **SC-001**: `docker compose up` brings graph-service + Neo4j; health returns ok.
- **SC-002**: Episode ingest + search integration test passes when Neo4j + OpenAI available.
- **SC-003**: JSON responses contain no Graphiti-specific field names (`EntityEdge`, `group_id` as Graphiti concept, etc.).
- **SC-004**: Go `go test ./...` and graph-service pytest green in CI.
- **SC-005**: Lore create path in Go does NOT call graph-service (deferred to F107).

## Assumptions

- Graphiti requires OpenAI API key for entity extraction; local dev uses env var.
- Neo4j 5 from existing docker-compose is sufficient.
- Scope maps to Graphiti `group_id` as `{kind}:{key}`.
- Full supersede/invalidation deferred; stub endpoints not required in v1 slice.

## Out of Scope

- F107 transactional outbox + async worker
- F108 full retrieval orchestration
- F109 `memlore.get_for_task`
- MCP tools exposing Graphiti/Neo4j
- OIDC / multi-tenant auth
