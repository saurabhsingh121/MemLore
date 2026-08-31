# Quickstart: Git Commit Ingestion (F030)

1. Create a local git repo with: (a) a commit whose body says why a migration
   was required (`because` / `migration`), (b) a `chore: bump version` commit,
   (c) a merge commit. Bind it to repository scope `github.com/acme/payments`.
2. `memlore ingest git --repository github.com/acme/payments --path <clone> --actor alice`
   — expect a succeeded run and exactly one candidate.
3. Repeat the same command — candidate count unchanged (idempotent).
4. `memlore ingest status --repository github.com/acme/payments` prints
   succeeded, counts, and cursor.
5. `POST /v1/ingest/git` with the same scope+path (local mode
   `X-Memlore-Actor`) returns the run; `GET /v1/ingest/candidates` lists the
   observational unverified lore with `evidence.type=commit`.
6. `POST /v1/lore-entries` still creates `human_authored` lore.
7. Compile context for a matching task with a **verified architecture**
   statement plus the git observation: architecture ranks first.
8. OIDC membership: a user without membership on that repository gets 403 on
   trigger and list.
9. MCP tool list still has 10 tools.
10. `go test ./...` and `go vet ./...` green.
