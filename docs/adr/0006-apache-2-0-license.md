# ADR 0006: Project license is Apache-2.0

- **Status**: Accepted
- **Date**: 2026-09-01

## Context

MemLore is open-source infrastructure (Go core + graph-service) intended for
adoption by humans and AI coding agents inside engineering orgs. The README
and `graph-service` already indicated Apache-2.0 as planned; a root license
file was still missing.

## Decision

Distribute MemLore under the **Apache License, Version 2.0**.

- Root `LICENSE` contains the full Apache-2.0 text.
- Root `NOTICE` carries the copyright attribution line.
- Public docs and badges state Apache-2.0 (not “planned”).

## Alternatives Considered

- **MIT**: simpler text; weaker explicit patent grant — rejected for infra.
- **GPL / AGPL**: stronger copyleft; harms enterprise/agent-host adoption.
- **BSL / SSPL**: protects against cloud freeloading; deferred unless that
  becomes a concrete product requirement.

## Consequences

- Contributions are assumed Apache-2.0 unless stated otherwise (Apache §5).
- Third-party deps must remain license-compatible (constitution dependency
  policy).
- Changing license later requires a superseding ADR and contributor consent
  for existing code.

## References

- [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0)
- `LICENSE`, `NOTICE`
- `graph-service/pyproject.toml` (`license = Apache-2.0`)
