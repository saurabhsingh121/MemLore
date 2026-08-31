# Architecture overview

MemLore is the engineering intelligence layer for humans and coding agents:
not a generic memory store, but governed knowledge of what the team decided,
why, whether it is still true, and what an agent needs before changing code.

It separates **governance** (who may know what, verification, audit) from
**knowledge** (temporal facts, relationships, semantic retrieval).

```text
                 Humans / Coding Agents
                          |
                     MCP / REST
                          |
                    MemLore Core
                        Go
                          |
        +-----------------+-----------------+
        |                                   |
        v                                   v
   PostgreSQL                         Graph Knowledge
 Control / Governance                    Service
       Plane                             Python
                                           |
                                        Graphiti
                                           |
                                         Neo4j
```

See [target-architecture.md](target-architecture.md) for full target design and
[current vs target](#current-vs-target) below.

## Planes

| Plane | Store | Owns |
|-------|-------|------|
| Governance / control | PostgreSQL | Users, teams, projects, repositories, scopes, permissions, authority metadata, verification, audit, ingestion state, transactional outbox |
| Knowledge | Graphiti + Neo4j | Semantic knowledge, graph relationships, temporal facts, episodes, graph retrieval |

Do **not** use distributed transactions across planes. Synchronize with a
transactional outbox (or equivalent reliable async mechanism).

## Current vs target

| Layer | Today | Target |
|-------|-------|--------|
| Core runtime | **Go** MemLore Core ([ADR-0005](../adr/0005-go-memlore-core.md)) | Go (unchanged) |
| REST / MCP | Go `memlore serve` / `memlore mcp` | Go adapters |
| Governance DB | PostgreSQL ✓ | PostgreSQL ✓ |
| Graph service | Thin Python `graph-service/` | Same |
| Outbox / workers | Go `memlore worker` | Go worker |

**Current delivery (v0.8.0 foundation)**: governance-plane lore (CRUD, verify,
invalidate, supersede, audit), dual-plane search, context compile, authority
evaluation, and optional OIDC/RBAC plus membership ACL on PostgreSQL via Go
REST `/v1/*` and ten `memlore.*` MCP tools. Knowledge-plane Graphiti/Neo4j is
isolated in `graph-service/`.

**Product direction**: capture (git/PR/ADR) → trust (review queue) → retrieve
(`get_for_task` / `why`) → use (coding agent) → detect drift (PR check) →
human correction → learn. See
[FEATURE_DEVELOPMENT.md](../development/FEATURE_DEVELOPMENT.md). Next specify:
F032 ADR ingest or F035 review queue, or F022 packet profiles. P0 human queues are CLI + REST; a
product web UI is F120 (P2).

## Adapters

- **MCP** — primary agent-facing contract (`memlore.*` tools)
- **REST** — UI, automation, integrations
- **CLI** — local developer workflows

## Related docs

- [Target architecture](target-architecture.md)
- [System context](system-context.md)
- [Containers](containers.md)
- [Authority model](authority-model.md)
- [Security](security.md)
- [Product feature roadmap](../development/FEATURE_DEVELOPMENT.md)
- [ADRs](../adr/)
