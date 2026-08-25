# ADR 0002: Python 3.12 + FastAPI application stack

- **Status**: Accepted
- **Date**: 2026-08-25

## Context

Graphiti is Python-native. MemLore needs REST, MCP, Pydantic validation, and
async I/O for API and workers.

## Decision

Use **Python 3.12**, **FastAPI**, **Pydantic**, **SQLAlchemy/Alembic**,
**pytest**, and the official MCP Python SDK as the primary application stack.
Package layout follows hexagonal boundaries under `src/memlore/`.

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
- `pyproject.toml`
