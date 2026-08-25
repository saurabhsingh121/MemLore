# Quickstart: Scoped Lore Entry (local)

## Prerequisites

- Docker Compose (Postgres)
- `uv` + Python 3.12
- Repo on branch `001-scoped-lore-entry`

## Steps (once implemented)

```bash
cp .env.example .env
docker compose up -d postgres
uv sync
uv run alembic upgrade head
uv run memlore serve
```

### Create

```bash
curl -sS -X POST http://127.0.0.1:8000/v1/lore-entries \
  -H 'Content-Type: application/json' \
  -H 'X-Memlore-Actor: alice' \
  -d '{
    "statement": "Payment events must use the transactional outbox.",
    "scope": {"kind": "repository", "key": "github.com/acme/payments"},
    "evidence": [{"type": "adr", "value": "0001-dual-plane-architecture"}]
  }'
```

### Get / Verify / Audits / List

```bash
ID=<id-from-create>

curl -sS http://127.0.0.1:8000/v1/lore-entries/$ID

curl -sS -X POST http://127.0.0.1:8000/v1/lore-entries/$ID/verify \
  -H 'X-Memlore-Actor: alice'

curl -sS http://127.0.0.1:8000/v1/lore-entries/$ID/audits

curl -sS 'http://127.0.0.1:8000/v1/lore-entries?scope_kind=repository&scope_key=github.com/acme/payments'
```

## Verify acceptance locally

```bash
uv run pytest tests/unit tests/contract tests/integration -q
```

Expect create→get→verify→audits→list behaviors from
`contracts/rest-lore-entries.md`.
