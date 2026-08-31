# Migrations

PostgreSQL schema changes MUST use **goose** migrations under `migrations/`,
applied with `memlore migrate` (embedded in the Go binary). Migrations must be
reviewable, deterministic, and safe.

## Local usage

```bash
docker compose up -d postgres
./scripts/install-memlore.sh
./bin/memlore migrate
```

DSN comes from `MEMLORE_POSTGRES_DSN` (see `.env.example`). Use a standard
`postgresql://` URL. A legacy `postgresql+psycopg://` prefix is still stripped
in Go for leftover env files.
