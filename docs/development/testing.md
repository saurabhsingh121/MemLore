# Testing

Test layout:

```text
tests/
├── unit/          # domain and pure application behavior
├── integration/   # Postgres, Neo4j, Graphiti, migrations, outbox
├── contract/      # REST and MCP contracts
└── e2e/           # sparse end-to-end journeys
```

Prefer many unit tests, fewer integration tests, and only necessary e2e tests.

Run:

```bash
uv run pytest
uv run pytest tests/unit -q
```
