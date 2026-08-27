# Implementation Plan: Context Compiler (F109)

**Branch**: `012-context-compiler` | **Date**: 2026-08-27 | **Spec**: [spec.md](./spec.md)

## Summary

Go context compiler atop `SearchKnowledgeHandler`: authority scoring, dedup,
ranking, token budgeting → ContextPacket. REST `POST /v1/context/compile` and
MCP `memlore.get_for_task`.

## Technical Context

**Language**: Go 1.25+ | **Deps**: F108 SearchKnowledgeHandler, presenters  
**Testing**: unit (ranking), handler tests, REST/MCP contracts

## Project Structure

```text
internal/application/context/ranking.go
internal/application/queries/compile_context.go
internal/adapters/presenters/context_packet.go
internal/adapters/http/handlers.go
internal/adapters/mcp/tools.go
specs/012-context-compiler/contracts/context-compile.md
```

## Phase 0 — Research

See [research.md](./research.md).

## Phase 1 — Design

- [data-model.md](./data-model.md)
- [contracts/context-compile.md](./contracts/context-compile.md)
