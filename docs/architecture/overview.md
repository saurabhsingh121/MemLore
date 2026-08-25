# Architecture overview

MemLore is a shared engineering-context platform for humans and coding agents.

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
| Core runtime | Python (FastAPI) | **Go** MemLore Core ([ADR-0005](../adr/0005-go-memlore-core.md)) |
| REST / MCP | Python adapters | Go adapters (strangler migration) |
| Governance DB | PostgreSQL ✓ | PostgreSQL ✓ |
| Graph service | Not built | Thin Python service → Graphiti → Neo4j |
| Outbox / workers | Not built | Go worker |

**Current delivery**: governance-plane lore CRUD/verify/audit on PostgreSQL via
Python REST `/v1/lore-entries` and MCP tools `memlore.remember`, `memlore.get`,
`memlore.verify`, `memlore.explain`, `memlore.search` (`uv run memlore mcp`).
Knowledge-plane Graphiti/Neo4j sync is not wired. Go skeleton (F101) is specified
on branch `003-go-core-skeleton`.

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
- [ADRs](../adr/)
