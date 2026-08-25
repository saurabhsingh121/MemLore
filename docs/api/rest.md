# REST API

Additional REST resources for lore entries, verification, listing, and audits
are specified in
[`specs/001-scoped-lore-entry/contracts/rest-lore-entries.md`](../../specs/001-scoped-lore-entry/contracts/rest-lore-entries.md)
(feature branch `001-scoped-lore-entry`; implementation pending).

| Method | Path | Purpose |
|--------|------|---------|
| `GET` | `/health` | Liveness payload (`status`, `service`, `version`) |
| `POST` | `/v1/lore-entries` | Create human-authored lore (planned) |
| `GET` | `/v1/lore-entries/{id}` | Get by id (planned) |
| `POST` | `/v1/lore-entries/{id}/verify` | Verify (planned) |
| `GET` | `/v1/lore-entries` | List by scope (planned) |
| `GET` | `/v1/lore-entries/{id}/audits` | List audits (planned) |
