# Research: Graph Retrieval Orchestration (F108)

## Parallel retrieval

**Decision**: Use `golang.org/x/sync/errgroup` to fetch governance and graph in parallel.

**Rationale**: Matches target architecture (parallel retrieval slice). Scoped context
cancellation propagates to both paths.

**Alternatives considered**: Sequential fetch (simpler but slower); manual goroutines
(errgroup is idiomatic and handles wait/cancel cleanly).

## Limit handling

**Decision**: Apply the same `limit` to each plane independently (default 10).

**Rationale**: Governance `ListByScope` returns all matches sorted by `created_at`
(newest first); graph-service enforces limit on search. Splitting limit across planes
would complicate v1 without clear product benefit.

**Alternatives considered**: Split limit 50/50 (rejected — arbitrary); single combined
ranking (deferred to F109).

## Endpoint naming

**Decision**: `POST /v1/knowledge-search` (not `/v1/search`).

**Rationale**: Avoids collision with lore CRUD list route semantics; name signals
dual-plane knowledge retrieval.

## Graph degradation

**Decision**: Log graph errors with `slog`; return `warnings: ["graph_service_unavailable"]`;
governance path errors still fail the request.

**Rationale**: Postgres is governance SoT; agents should get scope lore even when
graph-service is temporarily unavailable.

## Cross-plane deduplication

**Decision**: Deferred in v1.

**Rationale**: Governance entries and graph facts use different identifiers; dedup
requires authority enrichment (F109+).
