# Containers

| Container | Responsibility | Technology (baseline) |
|-----------|----------------|------------------------|
| Context API | REST, validation, authz boundary, transactions | FastAPI / Python 3.12 |
| MCP Gateway | Agent-facing domain tools | Official MCP Python SDK |
| Context Compiler | Retrieval → authority → conflicts → budgeted packet | Application module |
| Authority Engine | Explicit trust factors, verification, conflict policy | Domain + PostgreSQL |
| Ingestion Worker | Normalize sources, Graphiti episodes, retries | Redis + worker (Dramatiq planned) |
| Graphiti Core | Temporal graph knowledge engine | graphiti-core |
| Graph Store | Entities, facts, embeddings, indexes | Neo4j 5.x |
| Control Store | Identity, scope, policy, provenance, audit | PostgreSQL 16+ |
| Telemetry | Traces, metrics, logs | OpenTelemetry |

Local development runs Postgres, Neo4j, and Redis via `docker compose`.
