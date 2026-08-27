# Feature Specification: Transactional Outbox + Graph Sync (F107)

**Feature Branch**: `010-transactional-outbox`  
**Created**: 2026-08-26  
**Status**: Ready  
**Depends on**: F106 (graph-service), F104 (Go lore create)  
**Implements**: Product F004 (transactional outbox + graph sync)

## Goal

When lore is created on the governance plane, record an outbox event in the same
PostgreSQL transaction. A Go worker asynchronously publishes episodes to
graph-service — no synchronous dual-write and no distributed transactions.

## User Scenarios & Testing

### User Story 1 — Atomic lore + outbox write (Priority: P1)

**Given** a lore create request, **When** the handler commits, **Then** lore
and a pending outbox event exist together or neither exists.

**Independent Test**: Unit test with memory UoW proves outbox row on create.

### User Story 2 — Worker publishes episode (Priority: P1)

**Given** a pending outbox event, **When** the worker runs, **Then** graph-service
receives `POST /episodes` and the event is marked completed.

**Independent Test**: Worker test with fake KnowledgeGraph proves ingest called once.

### User Story 3 — Idempotent retry (Priority: P2)

**Given** a completed outbox event, **When** the worker runs again, **Then** no
duplicate graph ingest for the same lore id.

## Requirements

- **FR-001**: `outbox_events` goose migration
- **FR-002**: `CreateLore` adds `episode.ingest` outbox event in same UoW transaction
- **FR-003**: Go worker (`memlore worker`) polls pending events and calls `KnowledgeGraph.IngestEpisode`
- **FR-004**: Idempotency key = lore entry id for episode ingest
- **FR-005**: Mark completed / failed with attempts and last_error
- **FR-006**: Integration test with Postgres (skip if unavailable)

## Out of Scope

- F108 retrieval orchestration (read path)
- Redis / distributed locking beyond Postgres `SKIP LOCKED`
- Verify/supersede outbox events (create only in v1)

## Success Criteria

- **SC-001**: Lore create + outbox in one transaction (test proves)
- **SC-002**: Worker processes pending event with fake graph client
- **SC-003**: Re-processing completed event is safe (idempotent)
- **SC-004**: `go test ./...` green; migrate applies 00002
