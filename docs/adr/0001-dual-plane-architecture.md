# ADR 0001: Dual-plane architecture (PostgreSQL + Graphiti/Neo4j)

- **Status**: Accepted
- **Date**: 2026-08-25

## Context

MemLore must preserve governance concerns (identity, scopes, permissions,
verification, audit) and temporal semantic knowledge (facts, relationships,
episodes, retrieval). Mixing both in a single store either weakens
transactional governance or underuses graph/temporal capabilities.

## Decision

Use two planes:

1. **Governance plane** — PostgreSQL is the source of truth for users, teams,
   repositories, scopes, authority metadata, verification, audit, ingestion
   state, and transactional outbox.
2. **Knowledge plane** — Graphiti on Neo4j owns semantic knowledge, temporal
   facts, episodes, relationships, and graph retrieval.

Synchronize asynchronously via transactional outbox (or equivalent). Do not
require distributed transactions across planes.

## Alternatives Considered

- **Postgres-only**: insufficient native graph/temporal retrieval for the
  product vision.
- **Neo4j-only**: weak fit for RBAC, audit, and transactional governance.
- **Synchronous dual-write**: distributed transaction complexity and failure
  modes are unacceptable for v1.

## Consequences

- Application services must treat Postgres as authoritative for governance
  decisions even when graph retrieval returns candidates.
- Outbox/worker reliability becomes a core operational concern.
- Local Docker Compose must provide both Postgres and Neo4j.

## References

- `docs/architecture/overview.md`
- Source design: `docs/reference/engineering_context_platform_architecture.docx`
