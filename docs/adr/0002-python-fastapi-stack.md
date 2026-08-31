# ADR 0002: Python 3.12 + FastAPI application stack

- **Status**: Superseded (for MemLore Core) — see [ADR 0005](0005-go-memlore-core.md)
- **Date**: 2026-08-25

## Context

Graphiti is Python-native. MemLore needs REST, MCP, Pydantic validation, and
async I/O for API and workers.

## Decision

Use **Python 3.12**, **FastAPI**, **Pydantic**, **SQLAlchemy/Alembic**,
**pytest**, and the official MCP Python SDK as the primary application stack.
Package layout follows hexagonal boundaries under `src/memlore/`.

> **Update (2026-08-31)**: Superseded for the **MemLore Core control plane** by
> [ADR 0005](0005-go-memlore-core.md). Python remains the stack for the **graph
> knowledge service** (`graph-service/`). The legacy governance app
> (`src/memlore/`) was removed after Go slices verified (F106a remainder /
> `017-retire-python-core`).

## Alternatives Considered

- **Separate Java/Go gateway**: extra runtime and duplicated models without
  clear benefit while Graphiti integration remains Python-centric.
- **Litestar**: viable; FastAPI chosen for ecosystem familiarity and OpenAPI
  maturity for this team.

## Consequences

- One runtime for API, MCP adapter, and Graphiti integration.
- Strict typing (`mypy --strict`) and Ruff are part of CI.
- Domain code must not import FastAPI/SQLAlchemy/Neo4j directly.

## References

- Constitution: Architecture & Technology Baseline
- `graph-service/pyproject.toml`
