# REST API

| Method | Path | Purpose |
|--------|------|---------|
| `GET` | `/health` | Liveness payload (`status`, `service`, `version`) |
| `POST` | `/v1/lore-entries` | Create human-authored lore (`X-Memlore-Actor` required) |
| `GET` | `/v1/lore-entries/{id}` | Get by id |
| `POST` | `/v1/lore-entries/{id}/verify` | Verify (`X-Memlore-Actor` required; idempotent) |
| `GET` | `/v1/lore-entries` | List by `scope_kind` + `scope_key` |
| `GET` | `/v1/lore-entries/{id}/audits` | List audits (404 if entry missing) |
| `POST` | `/v1/knowledge-search` | Dual-plane knowledge search (governance + graph) |

Contract details:
[`specs/001-scoped-lore-entry/contracts/rest-lore-entries.md`](../../specs/001-scoped-lore-entry/contracts/rest-lore-entries.md).
Knowledge search: [`specs/011-graph-retrieval-orchestration/contracts/knowledge-search.md`](../../specs/011-graph-retrieval-orchestration/contracts/knowledge-search.md).

Example create:

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
