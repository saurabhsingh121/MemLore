# MemLore

Shared engineering memory for humans and AI coding agents.

MemLore preserves the *why* behind the code: architectural decisions, engineering
context, provenance, authority, temporal history, conflicts, superseded
knowledge, and agent observations — so coding agents retrieve trustworthy
context with evidence, not just similar text.

## Why MemLore?

Generic AI memory stores embeddings and snippets. Engineering work needs more:

- **Provenance** — who said it, human or agent, with evidence
- **Authority** — verified architecture outranks unverified inference
- **Temporal truth** — supersede facts without erasing history
- **Conflicts** — surface drift between policy and implementation
- **Scopes** — organization → team → project → repository → task

## Quick start

Requirements: Python 3.12+, [uv](https://docs.astral.sh/uv/), Docker.

```bash
cp .env.example .env
docker compose up -d
uv sync
uv run pytest
uv run memlore serve
```

Health check: `GET http://127.0.0.1:8000/health`

## Connect a coding agent

MCP tools are not shipped yet. The planned surface is intentionally small
(`memlore.get_for_task`, `memlore.search`, `memlore.remember`, …). See
[docs/api/mcp.md](docs/api/mcp.md).

## CLI

```bash
uv run memlore serve --host 127.0.0.1 --port 8000
```

Additional commands (`remember`, `recall`, `verify`, …) will land with the
first product features.

## Documentation

| Area | Link |
|------|------|
| Architecture overview | [docs/architecture/overview.md](docs/architecture/overview.md) |
| ADRs | [docs/adr/](docs/adr/) |
| Development setup | [docs/development/setup.md](docs/development/setup.md) |
| TDD | [docs/development/tdd.md](docs/development/tdd.md) |
| Constitution | [.specify/memory/constitution.md](.specify/memory/constitution.md) |

## Status

First product slice implemented: scoped human-authored lore entries
(store / retrieve / verify / list / audit) on PostgreSQL via REST `/v1/lore-entries`.
MCP, Graphiti sync, OIDC/RBAC, conflicts, and supersession remain future work.

## License

Apache-2.0 (planned; confirm before public release).
