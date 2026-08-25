# Tasks: Graph Knowledge Service (F106)

**Branch**: `009-graph-service` | **Spec**: [spec.md](./spec.md)

## Phase 1 — Setup

- [x] T001 Create `specs/009-graph-service/` spec, plan, contracts
- [x] T002 Scaffold `graph-service/` pyproject (uv, ruff, mypy, pytest)
- [x] T003 FastAPI app factory + `GET /health`
- [x] T004 Wire `graph-service` in `docker-compose.yml` + Dockerfile

## Phase 2 — Graphiti adapter + episodes (P1)

- [x] T005 RED: unit tests for episode request/response mapping
- [x] T006 GREEN: `graph_service/adapters/graphiti/` + `POST /episodes`
- [x] T007 Integration test: episode ingest round-trip (skip if Neo4j/OpenAI unavailable)

## Phase 3 — Search (P2)

- [x] T008 RED: unit tests for search mapping (no Graphiti field leakage)
- [x] T009 GREEN: `POST /search` with MemLore-shaped results
- [x] T010 Contract test: JSON schema keys validation

## Phase 4 — Facts stub + Go boundary (P2)

- [x] T011 `GET /facts/{id}` stub (404 or minimal retrieve)
- [x] T012 Go `KnowledgeGraph` port in `internal/application/ports/`
- [x] T013 Go HTTP client in `internal/infrastructure/graphclient/`
- [x] T014 Go integration contract test (health + ingest)

## Phase 5 — CI, docs, polish

- [x] T015 `graph-service` job in `.github/workflows/ci.yml`
- [x] T016 `graph-service/openapi.yaml` + `docs/api/graph-service.md`
- [x] T017 Update `FEATURE_DEVELOPMENT.md` + `.cursor/rules/specify-rules.mdc`
- [x] T018 Full test suite green (`go test ./...`, graph-service pytest)

## Dependencies

```text
T001 → T002 → T003 → T004
T005 → T006 → T007
T008 → T009 → T010
T012 → T013 → T014
T015–T018 after implementation complete
```

## Parallel opportunities

- T005/T008 unit tests can be written in parallel
- T012 Go port can start while Python search is in progress
