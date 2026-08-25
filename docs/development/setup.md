# Development setup

## Prerequisites

- Python 3.12+
- [uv](https://docs.astral.sh/uv/)
- Docker / Docker Compose

## Bootstrap

```bash
cp .env.example .env
docker compose up -d
uv sync
uv run pytest
uv run memlore serve
```

## Quality commands

```bash
uv run ruff check src tests
uv run ruff format src tests
uv run mypy
uv run pytest
```

## Spec Kit

This repo uses GitHub Spec Kit with the Cursor integration.

Typical feature flow:

1. `/speckit-constitution` (already ratified v1.0.0)
2. `/speckit-specify`
3. `/speckit-clarify` (if needed)
4. `/speckit-checklist`
5. `/speckit-plan`
6. `/speckit-tasks`
7. `/speckit-analyze`
8. `/speckit-implement`

Constitution: `.specify/memory/constitution.md`
