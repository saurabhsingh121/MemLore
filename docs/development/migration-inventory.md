# MemLore Migration Inventory

**Last Updated**: 2026-08-25  
**Purpose**: Track each capability from current implementation → target architecture.  
**Statuses**: `Not Started` · `Characterizing` · `Specified` · `In Development` · `Migrated` · `Verified` · `Deprecated` · `Removed` · `Blocked`

| Capability | Current Implementation | Target | Status | Tests | Migration Notes |
|------------|------------------------|--------|--------|-------|-----------------|
| **Governance plane** |
| Lore create (human-authored) | Python `CreateLoreHandler` + SQLAlchemy | Go command + sqlc | Not Started | Partial (unit, contract, integration*) | Characterize from `tests/unit/application/test_create_lore.py`; first Go slice |
| Lore get by id | Python `GetLoreHandler` | Go query + sqlc | Not Started | Partial | |
| Lore list by scope | Python `ListLoreByScopeHandler` | Go query + sqlc | Not Started | Partial | Exact kind+key match only |
| Lore verify | Python `VerifyLoreHandler` + domain `apply_verification` | Go command + domain | Not Started | Partial | Idempotent; self-verify allowed |
| Audit create/verify | Python handlers + `audit_records` table | Go + sqlc | Not Started | Partial | |
| Audit list by lore id | Python `ListAuditsHandler` | Go query | Not Started | Partial | 404 if lore missing |
| Domain validation (statement, scope, evidence) | Python dataclasses | Go domain package | Verified | Yes | F102 characterization tests |
| Health endpoint | Python FastAPI `/health` | Go chi handler | Not Started | Yes (unit) | Trivial; good Go smoke test |
| **API adapters** |
| REST `/v1/lore-entries` | FastAPI `routes_lore.py` | Go `adapters/http` | Not Started | Yes (contract) | Preserve JSON shapes during strangler |
| REST actor header | `X-Memlore-Actor` deps | Go middleware | Not Started | Yes (contract) | No OIDC yet |
| MCP `memlore.remember` | Python MCP adapter | Go MCP SDK | Not Started | Yes (unit, contract, e2e*) | Defer until REST Go slice green or parallel |
| MCP `memlore.get` | Python MCP adapter | Go MCP SDK | Not Started | Yes | |
| MCP `memlore.verify` | Python MCP adapter | Go MCP SDK | Not Started | Yes | |
| MCP `memlore.explain` | Python MCP adapter | Go MCP SDK | Not Started | Yes | Entry + audits, no NL narrative |
| MCP `memlore.search` | Python MCP adapter (scope list) | Go MCP SDK | Not Started | Yes | Not semantic search |
| MCP stdio CLI | `memlore mcp` Python | `memlore mcp` Go binary | Not Started | Yes (e2e*) | |
| CLI `memlore serve` | Uvicorn Python | Go HTTP server | Not Started | No | |
| **Persistence** |
| Postgres schema `lore_entries` | Alembic `0001` + goose `00001` | goose migration | Migrated | Partial | SQL ported; integration test build tag |
| Postgres schema `audit_records` | Alembic `0001` + goose `00001` | goose migration | Migrated | Partial | |
| Repository layer | SQLAlchemy repos | sqlc + pgx repos | Migrated | Yes | F103 integration tests |
| Unit of work / transactions | `SqlAlchemyUnitOfWork` | `postgres.BeginUnitOfWork` | Migrated | Yes | pgx transaction |
| In-memory test doubles | `tests/support/fakes.py` | Go fake repos | Not Started | Yes | For unit/contract without DB |
| **Knowledge plane** |
| Graphiti integration | None | Python `graph-service` | Not Started | No | Greenfield; no code to migrate |
| Neo4j connectivity | Docker only | Python service | Not Started | No | |
| Episode ingestion | None | `POST /episodes` graph-service | Not Started | No | After outbox |
| Semantic/graph search | None | `POST /search` graph-service | Not Started | No | Distinct from scope list |
| Fact get/invalidate/supersede | None | graph-service REST | Not Started | No | Spec required |
| **Orchestration** |
| Transactional outbox | None | Go outbox + worker | Not Started | No | ADR candidate |
| Background worker | None | `cmd/worker` Go | Not Started | No | Redis/Dramatiq was planned in docs only |
| Graph sync after lore write | None | Outbox → graph-service | Not Started | No | No dual-write |
| **Domain features (not built in Python)** |
| Authority factor persistence | Docs only | Go domain + Postgres | Not Started | No | Spec before implement |
| Authority scoring / ranking | Docs only | Go application service | Not Started | No | Explainable factors required |
| Verification canonical/invalidated | Enum partial | Go domain | Not Started | No | Extend `VerificationStatus` |
| Supersession | None | Go governance + graph coord | Not Started | No | |
| Conflict detection | None | Go application | Not Started | No | |
| Agent-authored lore origins | Enum reserved | Go create paths | Not Started | No | MCP remember currently human only |
| Users / teams / projects / repos | None | Go + Postgres | Not Started | No | |
| OIDC / RBAC | None | Go auth adapter | Not Started | No | |
| Context compiler | None | Go application | Not Started | No | |
| `memlore.get_for_task` | None | Go MCP | Not Started | No | |
| `memlore.supersede` / `invalidate` | None | Go MCP + graph | Not Started | No | |
| **Infrastructure** |
| Configuration | Pydantic settings | Go typed env config | Not Started | No | |
| Structured logging | Python `get_logger` | Go `slog` | Not Started | No | |
| OpenTelemetry | None | Go + Python OTel | Not Started | No | |
| Docker Compose app services | DBs only | Add memlore + graph-service | Not Started | No | |
| CI pipeline | Python ruff/mypy/pytest | Add Go jobs | Migrated | N/A | `go-test` job in CI |
| **Documentation / process** |
| Spec Kit workflow | Active | Continue | Migrated | N/A | |
| Go module skeleton | — | `go.mod`, layout, `cmd/memlore` | Verified | Yes | F101; `go test ./...` |
| ADR: Go core | ADR-0005 | ADR-0005 | Migrated | N/A | Accepted |
| ADR: Graphiti isolation | Partial (0001, 0003) | ADR-0002 target variant | Not Started | N/A | |
| Migration discovery | This effort | Maintained | Verified | N/A | `MIGRATION_DISCOVERY.md` |

\* Integration and e2e tests require PostgreSQL (`docker compose up -d postgres`); they skip when DB is unavailable. CI currently does not start Postgres.

---

## Status Summary (2026-08-25)

| Status | Count |
|--------|------:|
| Verified (Go F101 skeleton) | Go module, goose DDL port, CI |
| Migrated (partial) | Postgres schema parity in goose |
| In production (Python) | F001+F002 governance REST/MCP |
| Not Started (Go features) | Lore handlers, MCP, graph-service |
| Blocked | 0 |

---

## Next Inventory Updates

Update this table when:

1. Characterization tests exist for a capability (`Characterizing` → `Specified`)
2. Spec Kit spec + plan approved (`Specified`)
3. Go implementation begins (`In Development`)
4. Contract tests pass against Go (`Migrated`)
5. Integration + parity checks pass (`Verified`)
6. Python path removed (`Removed`)

---

## Related Documents

- [MIGRATION_DISCOVERY.md](MIGRATION_DISCOVERY.md)
- [FEATURE_DEVELOPMENT.md](FEATURE_DEVELOPMENT.md)
