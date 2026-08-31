# ADR 0004: Canonical project brand is MemLore

- **Status**: Accepted
- **Date**: 2026-08-25
- **Amended**: 2026-09-01 — removed alternate-brand references from the tree

## Context

Package, CLI, MCP namespace, and documentation must share one canonical public
name so agents and humans do not encounter conflicting product identity.

## Decision

Use **MemLore** as the canonical project brand:

- Repository / product: MemLore
- Package / CLI: `memlore`
- MCP namespace: `memlore.*`
- Domain metaphor “lore” remains valid vocabulary (lore entry, team lore, …)

No alternate product brand is used in code, contracts, or public docs.

## Alternatives Considered

- **Dual branding**: creates irreversible naming drift across CLI/MCP/docs.
- **Rename later without an ADR**: high coordination cost across adapters and
  documentation; rejected in favor of an explicit brand decision.

## Consequences

- All bootstrap artifacts and brand assets use MemLore.
- Brand assets live under `docs/brand/`.
- Renaming later requires a superseding ADR and coordinated contract changes.

## References

- `.specify/memory/constitution.md` (Architecture & Technology Baseline)
- `docs/brand/README.md`
