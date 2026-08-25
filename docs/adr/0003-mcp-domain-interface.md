# ADR 0003: Domain MCP interface over raw Graphiti tools

- **Status**: Accepted
- **Date**: 2026-08-25

## Context

Coding agents need a stable, small contract. Exposing Graphiti internals couples
clients to storage details and skips authority, scope, and conflict handling.

## Decision

Expose a MemLore domain MCP surface. Preferred initial operations:

- `memlore.get_for_task`
- `memlore.search`
- `memlore.remember`
- `memlore.get`
- `memlore.verify`
- `memlore.supersede`
- `memlore.invalidate`
- `memlore.explain`

`get_for_task` should become the preferred coding-agent operation and return a
compiled context packet, not a raw vector-search dump.

## Alternatives Considered

- **Raw Graphiti MCP in production**: rejected for product boundary and
  authority/safety reasons (allowed only as a debug escape hatch if ever needed).
- **REST-only for agents**: weaker fit for MCP-native coding agents.

## Consequences

- MCP tools are product API; changes need contract tests and docs.
- Context compiler is a first-class subsystem.
- Graphiti remains an infrastructure detail behind ports.

## References

- `docs/api/mcp.md`
- Constitution principles III, V, VI
