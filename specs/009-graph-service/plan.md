# Implementation Plan: Graph Knowledge Service (F106)

**Branch**: `009-graph-service` | **Date**: 2026-08-25 | **Spec**: [spec.md](./spec.md)  
**ADR**: [ADR-0005](../../docs/adr/0005-go-memlore-core.md), [ADR-0001](../../docs/adr/0001-dual-plane-architecture.md)

## Summary

Greenfield Python `graph-service/` FastAPI app isolating Graphiti behind MemLore
HTTP contracts. Go defines `KnowledgeGraph` port + HTTP client with contract
tests. No wiring to lore create path.

## Technical Context

**Language/Version**: Python 3.12 (graph-service), Go 1.25+ (client port)  
**Primary Dependencies**: FastAPI, Pydantic, graphiti-core, neo4j, uvicorn; Go stdlib HTTP  
**Storage**: Neo4j 5 via Graphiti (knowledge plane only)  
**Testing**: pytest (unit/integration/contract), `go test` + `integration` build tag  
**Target Platform**: Docker Compose local dev; GitHub Actions CI  
**Project Type**: Monorepo — `graph-service/` sibling to Go `internal/`  
**Performance Goals**: Health < 500ms; search P95 not benchmarked in v1  
**Constraints**: No Graphiti types in Go; PostgreSQL remains governance SoT  
**Scale/Scope**: Vertical slice — health, episodes, search, facts stub, Go client

## Constitution Check

- [x] TDD: RED → GREEN for episodes, search, Go contract tests
- [x] Spec-driven: spec.md FR + SC defined
- [x] Architecture integrity: dual-plane boundary; Graphiti behind adapter
- [x] Documentation: OpenAPI, graph-service.md, FEATURE_DEVELOPMENT update
- [x] Authority & provenance: provenance_refs on episode ingest (metadata only in v1)
- [x] Temporal correctness: Graphiti bi-temporal model delegated to adapter
- [x] Secure by default: no MCP exposure of graph internals
- [x] Observability: structured logging in graph-service (stdlib logging)
- [x] Dependency policy: graphiti-core justified in research.md
- [x] Simplicity: no outbox, no Go orchestration in F106

## Project Structure

```text
graph-service/
├── pyproject.toml
├── openapi.yaml
├── Dockerfile
├── src/graph_service/
│   ├── adapters/
│   │   ├── http/          # FastAPI routes, schemas
│   │   └── graphiti/      # Graphiti adapter (only place for graphiti imports)
│   ├── application/
│   │   └── ports/         # KnowledgeGraph protocol
│   └── bootstrap/         # settings, app factory
└── tests/
    ├── unit/
    ├── integration/
    └── contract/

internal/application/ports/knowledge_graph.go
internal/infrastructure/graphclient/
specs/009-graph-service/contracts/graph-service-api.md
docs/api/graph-service.md
```

## Phase 0 — Research

See [research.md](./research.md).

## Phase 1 — Design

- [data-model.md](./data-model.md)
- [contracts/graph-service-api.md](./contracts/graph-service-api.md)
- [quickstart.md](./quickstart.md)
