# Research: Retire Python core

## R1 — Delete vs keep as characterization library

**Decision**: Delete `src/memlore/` entirely. Keep Go comments that cite
former Python paths.

**Rationale**: Python core is a stale five-tool server on Alembic-0001 only.
Keeping it “for characterization” invites running the wrong binary.

## R2 — Root pytest

**Decision**: Delete root `tests/`. Graph-service tests stay under
`graph-service/tests/`.

## R3 — CI

**Decision**: Remove the `quality` job. Keep `go-test`, `go-integration`,
`graph-service`.

## R4 — Constitution version

**Decision**: Amend to **1.1.0** (MINOR). Principles I–VIII unchanged.
Architecture baseline and workflow (goose, layout) catch up to ADR-0005 now
that the strangler is complete. Not MAJOR: ADR-0005 already superseded
Python for core; this records completion.

## R5 — Historical specs

**Decision**: Do not rewrite `specs/001`–`016`. They are the build record.

## R6 — DSN

**Decision**: `.env.example` uses `postgresql://`. Keep Go
`postgresql+psycopg://` strip for leftover env files.
