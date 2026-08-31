# Feature Specification: Context Compiler + get_for_task (F109)

**Feature Branch**: `012-context-compiler`  
**Created**: 2026-08-27  
**Status**: Ready  
**Depends on**: F108 (graph retrieval orchestration)  
**Implements**: Product F007 (partial — context compiler v1)

## Goal

Compile a token-budgeted **ContextPacket** for coding agents by retrieving
knowledge (via F108), applying authority enrichment and ranking, deduplicating
cross-plane hits, and exposing the result via REST and `memlore.get_for_task`.

## User Scenarios & Testing

### User Story 1 — Compiled context for a task (Priority: P1)

**Given** a task description, search query, and scope, **When** the agent calls
get_for_task, **Then** a ranked, budgeted list of context items is returned with
authority metadata.

**Independent Test**: Unit test with fake search results proves ranked items and
token meta.

### User Story 2 — Authority ranking (Priority: P1)

**Given** verified governance lore and graph facts, **When** context is compiled,
**Then** verified governance ranks above unverified and graph-only hits include
explainable authority factors.

**Independent Test**: Unit test on ranking functions.

### User Story 3 — REST and MCP parity (Priority: P2)

**Given** the same inputs, **When** REST or MCP is called, **Then** response
shape matches the contract.

**Independent Test**: Contract tests.

## Requirements

- **FR-001**: `CompileContextHandler` calls `SearchKnowledgeHandler` then compiles
- **FR-002**: Response includes `task`, `query`, `scope`, `items`, `meta`, `warnings`
- **FR-003**: Each item has `source` (`governance`|`graph`), `authority_score`,
  `authority_factors`, and `statement`
- **FR-004**: v1 authority: verification status + origin for governance; graph
  score for graph hits
- **FR-005**: v1 dedup: skip graph item when statement matches governance (normalized)
- **FR-006**: Token budgeting: estimate tokens per item; pack by rank until budget
- **FR-007**: `POST /v1/context/compile` REST endpoint
- **FR-008**: MCP `memlore.get_for_task` with same response shape
- **FR-009**: `memlore.knowledge_search` unchanged

## Out of Scope

- Full authority factor model (F003)
- File/branch scope resolution
- OIDC / multi-tenant auth

*(Conflict detection and supersede/invalidation filtering delivered in F112.)*

## Success Criteria

- **SC-001**: get_for_task returns ContextPacket with ranked items
- **SC-002**: Verified lore outranks graph-only hits in unit test
- **SC-003**: Token budget caps included items
- **SC-004**: REST and MCP contract tests pass
- **SC-005**: `go test ./...` green

## Assumptions

- `scope` required in v1 (matches F108 governance path)
- `query` defaults to `task` when omitted
- Default token budget 4096; char/4 token estimate
- F007 marked **partial** when this ships
