# Migrations

PostgreSQL schema changes MUST use Alembic migrations. Migrations must be
reviewable, deterministic, and safe. Prefer verifying upgrade and downgrade
paths.

## Local usage

```bash
docker compose up -d postgres
uv run alembic upgrade head
uv run alembic revision --autogenerate -m "describe change"
uv run alembic downgrade -1   # when rollback is supported
```

Config: `alembic.ini` + `alembic/env.py`. Migration scripts live in
`alembic/versions/`.

DSN comes from `MEMLORE_POSTGRES_DSN` (see `.env.example`).
