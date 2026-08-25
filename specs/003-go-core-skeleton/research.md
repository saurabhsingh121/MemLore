# Research: Go Core Project Skeleton (F101)

**Feature**: `003-go-core-skeleton`  
**Date**: 2026-08-25

## R1 — Go version

**Decision**: Go **1.25** minimum (`go.mod` directive `go 1.25`).

**Rationale**: Aligns with migration target architecture prompt; CI uses
`actions/setup-go` with `1.25.x`.

**Alternatives**: Go 1.24 — rejected (explicit 1.25+ target).

## R2 — Module path

**Decision**: `github.com/memlore/memlore`

**Rationale**: Matches `pyproject.toml` project URLs and README placeholders.

## R3 — HTTP router (future slices)

**Decision**: **chi** v5 — not wired in F101 except as documented dependency
optional stub.

**Rationale**: Lightweight, idiomatic, matches target architecture; no framework
in F101 HTTP surface required.

## R4 — PostgreSQL access

**Decision**: **pgx/v5** driver + **sqlc** for query generation.

**Rationale**: Explicit SQL visibility; constitution-aligned; no ORM domain
coupling.

**F101 scope**: sqlc.yaml + one smoke query (`SELECT 1` or schema introspection);
full lore queries land in F103.

## R5 — Migrations

**Decision**: **goose** v3 CLI; SQL files in `migrations/` at repo root (Go
convention).

**Rationale**: Matches target stack; Alembic remains for Python legacy until
Go path owns schema evolution.

**Port strategy**: Copy DDL from `alembic/versions/0001_lore_audit.py` to
`migrations/00001_lore_audit.sql` with goose annotations.

## R6 — Logging

**Decision**: **slog** stdlib in `cmd/memlore` stub.

**Rationale**: Target observability baseline; no third-party logger in F101.

## R7 — Testing

**Decision**: stdlib `testing` + table-driven tests; **testify** only if
assertions reduce noise (optional, prefer stdlib in F101).

**Integration**: Document `testcontainers-go` for F103; F101 may include a
build-tagged skipped integration test for goose apply.

## R8 — CI integration

**Decision**: Add `go-test` job to `.github/workflows/ci.yml` after Python job
(same workflow file, parallel jobs).

**Rationale**: Single pipeline visibility; no regression to Python gates.

## R9 — Coexistence with Python

**Decision**: Python `src/memlore/` unchanged; Go tree is additive.

**Rationale**: Strangler migration; F001/F002 remain operational.

## R10 — Directory creation rule

**Decision**: Only create packages with at least one `.go` file; no empty
`internal/domain/lore/` until F102.

**Rationale**: User migration prompt: "Do not create empty directories merely
to match structure."

## R11 — ADR linkage

**Decision**: ADR-0005 accepted before implementation; ADR-0002 marked
superseded for core runtime.

**Rationale**: Records irreversible stack choice per constitution.
