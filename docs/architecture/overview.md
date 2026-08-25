# Architecture overview

MemLore is a shared engineering-context platform for humans and coding agents.

It separates **governance** (who may know what, verification, audit) from
**knowledge** (temporal facts, relationships, semantic retrieval).

```text
                 Humans / Coding Agents
                         |
                +--------+--------+
                |                 |
               MCP               REST
                |                 |
                +--------+--------+
                         |
             Engineering Context Service
                         |
     +-------------------+-------------------+
     |                                       |
     v                                       v
Governance Plane                      Knowledge Plane
   PostgreSQL                        Graphiti + Neo4j
 users / teams / scopes              entities / facts
 authority / verification            episodes / temporal
 provenance / audit / outbox         semantic retrieval
```

## Planes

| Plane | Store | Owns |
|-------|-------|------|
| Governance / control | PostgreSQL | Users, teams, projects, repositories, scopes, permissions, authority metadata, verification, audit, ingestion state, transactional outbox |
| Knowledge | Graphiti + Neo4j | Semantic knowledge, graph relationships, temporal facts, episodes, graph retrieval |

Do **not** use distributed transactions across planes. Synchronize with a
transactional outbox (or equivalent reliable async mechanism).

## Adapters

- **MCP** — primary agent-facing contract (`memlore.*` tools)
- **REST** — UI, automation, integrations
- **CLI** — local developer workflows

## Related docs

- [System context](system-context.md)
- [Containers](containers.md)
- [Authority model](authority-model.md)
- [Security](security.md)
- [ADRs](../adr/)
