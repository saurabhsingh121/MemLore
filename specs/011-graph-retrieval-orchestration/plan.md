# Implementation Plan: Graph Retrieval Orchestration (F108)

**Branch**: `011-graph-retrieval-orchestration` | **Date**: 2026-08-27 | **Spec**: [spec.md](./spec.md)  
**ADR**: [ADR-0005](../../docs/adr/0005-go-memlore-core.md), [ADR-0001](../../docs/adr/0001-dual-plane-architecture.md), [ADR-0003](../../docs/adr/0003-mcp-domain-interface.md)

## Summary

Go application orchestrator parallel-fetches governance scope list (when scope
provided) and knowledge graph search via `ports.KnowledgeGraph`. REST
`POST /v1/knowledge-search` and MCP `memlore.knowledge_search` expose the merged
MemLore-shaped response. Graphiti stays behind graph-service; no write-path changes.

## Technical Context

**Language/Version**: Go 1.25+  
**Primary Dependencies**: chi, go-sdk/mcp, errgroup, existing graphclient  
**Storage**: PostgreSQL (governance read); graph-service HTTP (knowledge read)  
**Testing**: `go test` unit + contract; `integration` build tag for graph-service  
**Target Platform**: `memlore serve`, `memlore mcp`  
**Performance Goals**: Parallel fetch; no cross-plane ranking in v1  
**Constraints**: No Graphiti imports in Go; `memlore.search` contract frozen  
**Scale/Scope**: Read-path vertical slice only

## Constitution Check

- [x] TDD: RED → GREEN for orchestrator, REST, MCP contract tests
- [x] Spec-driven: spec.md FR + SC defined
- [x] Architecture integrity: dual-plane; application → ports only
- [x] Documentation: rest.md, mcp.md, FEATURE_DEVELOPMENT update
- [x] Secure by default: no Graphiti keys in API responses
- [x] Simplicity: minimal ranking; no dedup in v1

## Project Structure

```text
internal/application/queries/search_knowledge.go
internal/adapters/presenters/knowledge_search.go
internal/adapters/http/handlers.go          # POST /v1/knowledge-search
internal/adapters/mcp/tools.go              # memlore.knowledge_search
cmd/memlore/main.go                         # wire graphclient
specs/011-graph-retrieval-orchestration/contracts/knowledge-search.md
```

## Phase 0 — Research

See [research.md](./research.md).

## Phase 1 — Design

- [data-model.md](./data-model.md)
- [contracts/knowledge-search.md](./contracts/knowledge-search.md)
- [quickstart.md](./quickstart.md)
