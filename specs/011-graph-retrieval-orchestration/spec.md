# Feature Specification: Graph Retrieval Orchestration (F108)

**Feature Branch**: `011-graph-retrieval-orchestration`  
**Created**: 2026-08-27  
**Status**: Ready  
**Depends on**: F106 (graph-service), F107 (outbox sync)  
**Implements**: Product F006 (partial — semantic search + graph retrieval read path)

## Goal

Add a Go application-layer retrieval orchestrator that queries the knowledge plane
via `KnowledgeGraph.Search` (graph-service HTTP), optionally in parallel with
governance-plane scope listing, and returns a single MemLore-shaped response
suitable for REST and MCP adapters. `memlore.search` remains exact scope list only.

## User Scenarios & Testing

### User Story 1 — Parallel governance + graph search (Priority: P1)

**Given** a query and scope, **When** the client calls knowledge search,
**Then** governance lore entries and graph facts are returned in one response.

**Independent Test**: Unit test with fake graph + memory UoW proves both planes
called and merged shape.

**Acceptance Scenarios**:

1. **Given** lore in scope A and graph facts for query Q, **When** knowledge
   search runs with Q and scope A, **Then** `governance.items` and `graph.items`
   are both populated.
2. **Given** no scope in request, **When** knowledge search runs, **Then**
   `governance.items` is empty and graph search still runs.
3. **Given** empty graph results, **When** knowledge search runs, **Then**
   `graph.items` is an empty array (not null).

### User Story 2 — Graceful graph degradation (Priority: P1)

**Given** Postgres is healthy but graph-service is down, **When** knowledge search
runs with scope, **Then** governance results are returned with a warning.

**Independent Test**: Unit test with graph client returning error proves
`warnings` contains `graph_service_unavailable` and governance items present.

### User Story 3 — REST and MCP parity (Priority: P2)

**Given** the same inputs, **When** REST `POST /v1/knowledge-search` or MCP
`memlore.knowledge_search` is called, **Then** response JSON shape matches the
contract (no Graphiti-specific keys).

**Independent Test**: Contract tests for REST and MCP JSON shapes.

## Requirements

- **FR-001**: `SearchKnowledgeHandler` orchestrates parallel governance list (when
  scope provided) and `KnowledgeGraph.Search`
- **FR-002**: Response includes `query`, optional `scope`, `governance.items`,
  `graph.items`, and `warnings`
- **FR-003**: Graph results sorted by score descending; governance by `created_at`
  descending
- **FR-004**: Graph-service failure MUST NOT fail the entire request when
  governance succeeded; add `graph_service_unavailable` warning
- **FR-005**: `POST /v1/knowledge-search` REST endpoint with MemLore error envelope
- **FR-006**: MCP tool `memlore.knowledge_search` with same response shape
- **FR-007**: `memlore.search` behavior unchanged (exact scope list only)
- **FR-008**: No Graphiti types in Go; application depends on `ports.KnowledgeGraph` only
- **FR-009**: Integration test with real graph-service when available (skip otherwise)

## Out of Scope

- F109 `memlore.get_for_task` / full context compiler
- Changing `memlore.search` to semantic search
- Cross-plane deduplication (deferred v1)
- PostgreSQL schema changes
- OIDC / multi-tenant auth

## Success Criteria

- **SC-001**: `POST /v1/knowledge-search` returns governance + graph sections
- **SC-002**: `memlore.knowledge_search` returns same shape as REST
- **SC-003**: Unit test proves `KnowledgeGraph.Search` called with correct args
- **SC-004**: Graph down → governance results + warning (unit test)
- **SC-005**: `memlore.search` contract tests unchanged
- **SC-006**: `go test ./...` green

## Assumptions

- F006 is marked **partial** when this ships; full F006 may add scope-less
  governance search later.
- Same `limit` applies independently to each plane (governance list capped by
  repo query; graph search capped by graph-service).
- `actor_id` on MCP is validated but not used for authorization in v1.
