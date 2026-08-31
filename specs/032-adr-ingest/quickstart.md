# Quickstart: ADR Auto-Ingestion (F032)

1. Bind repository scope `github.com/acme/payments`. Tests use a fake
   `ADRReader` or `t.TempDir()` fixtures (no network).
2. Fixture tree under `--path`:
   - `docs/adr/0001-use-postgres.md` — Status: Accepted, Decision heading
   - `docs/adr/0002-draft.md` — Status: Draft
   - `docs/adr/README.md`
   - `docs/adr/template.md`
3. `memlore ingest adr --repository github.com/acme/payments --path <dir> --actor alice`
   — expect a succeeded run and exactly one current lore entry (verified
   `architecture_decision`, evidence type `adr`).
4. Repeat the same command — current lore count unchanged (idempotent).
5. Change the Decision text of `0001-use-postgres.md`, ingest again — new
   current lore; previous ingest-created entry is superseded, not deleted.
6. `memlore ingest status --repository github.com/acme/payments --kind adr`
   prints succeeded and counts. Default `ingest status` still shows git.
7. `POST /v1/ingest/adr` with the same scope and path (local mode
   `X-Memlore-Actor`) returns the run; `GET /v1/ingest/candidates?evidence_type=adr`
   lists the verified architecture-decision lore.
8. `POST /v1/lore-entries` still creates `human_authored` lore.
9. `POST /v1/ingest/git` and `POST /v1/ingest/pr` still work.
10. Compile context for a matching task with the ingested accepted ADR plus a
    git or PR observation: ADR-derived lore ranks first.
11. OIDC membership: a user without membership on that repository gets 403 on
    trigger and list.
12. MCP tool list still has 10 tools.
13. Missing `--path` directory yields a failed observable run and no lore.
14. `go test ./...` and `go vet ./...` green.
