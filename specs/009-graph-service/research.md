# Research: F106 Graph Service

## R1 — Graphiti package

**Decision**: Use `graphiti-core` from PyPI with Neo4j 5 driver.  
**Rationale**: Official Graphiti stack per target architecture and Neo4j blog.  
**Alternatives**: Direct Neo4j Cypher (rejects temporal/episode model).

## R2 — OpenAI requirement

**Decision**: Require `OPENAI_API_KEY` for Graphiti ingest/search in real integration;
unit tests mock the adapter; integration tests skip without key.  
**Rationale**: graphiti-core uses OpenAI for extraction and embeddings by default.

## R3 — Scope → group_id

**Decision**: Map MemLore scope to Graphiti `group_id` as `{kind}:{key}`.  
**Rationale**: Aligns tenant/repo isolation with governance scope model.

## R4 — Episode body shape

**Decision**: JSON episode body with `statement`, `metadata`, `provenance_refs`.  
**Rationale**: Preserves provenance without exposing Graphiti episode schema.

## R5 — Go client location

**Decision**: `internal/infrastructure/graphclient` implements `ports.KnowledgeGraph`.  
**Rationale**: Matches existing postgres/memory infrastructure pattern.

## R6 — No lore wiring

**Decision**: Do not call graph client from create_lore handler.  
**Rationale**: F107 outbox owns dual-write; avoids distributed transaction risk.
