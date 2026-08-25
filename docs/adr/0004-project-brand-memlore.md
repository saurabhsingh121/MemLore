# ADR 0004: Canonical project brand is MemLore

- **Status**: Accepted
- **Date**: 2026-08-25

## Context

Source materials included both **MemLore** (agent/product brief and repository
intent) and a recommended alternate brand **Tavryn**
(`docs/reference/tavryn_brand_cli_decision_guide.docx`). Package, CLI, MCP
namespace, and docs must share one canonical name.

## Decision

Use **MemLore** as the canonical project brand:

- Repository / product: MemLore
- Python package / CLI: `memlore`
- MCP namespace: `memlore.*`
- Domain metaphor “lore” remains valid vocabulary (lore entry, team lore, …)

Tavryn branding is **deferred** and not used in code or public contracts unless
a future ADR supersedes this decision.

## Alternatives Considered

- **Tavryn now**: stronger brand guide rationale, but conflicts with the
  established MemLore agent brief and current repository direction.
- **Dual branding**: creates irreversible naming drift across CLI/MCP/docs.

## Consequences

- All bootstrap artifacts use MemLore.
- Brand guide retained under `docs/reference/` for historical context only.
- Renaming later requires a superseding ADR and coordinated contract changes.

## References

- `.specify/memory/constitution.md` (Architecture & Technology Baseline)
- `docs/reference/tavryn_brand_cli_decision_guide.docx`
