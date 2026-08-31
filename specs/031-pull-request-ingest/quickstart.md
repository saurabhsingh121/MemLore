# Quickstart: Pull Request Ingestion (F031)

1. Bind repository scope `github.com/acme/payments`. Provide a GitHub token
   (`MEMLORE_GITHUB_TOKEN` or `GITHUB_TOKEN`). Tests use a fake
   `PullRequestReader` (no network).
2. Fixture PRs: (a) merged PR whose body says why a migration was required
   (`because` / `migration`), (b) merged dependabot bump with no review
   rationale, (c) open unmerged PR.
3. `memlore ingest pr --repository github.com/acme/payments --actor alice`
   — expect a succeeded run and exactly one candidate.
4. Repeat the same command — candidate count unchanged (idempotent).
5. `memlore ingest status --repository github.com/acme/payments --kind pr`
   prints succeeded, counts, and cursor. Default `ingest status` still shows git.
6. `POST /v1/ingest/pr` with the same scope (local mode `X-Memlore-Actor`)
   returns the run; `GET /v1/ingest/candidates?evidence_type=pr` lists the
   observational unverified lore with `evidence.type=pr`.
7. `POST /v1/lore-entries` still creates `human_authored` lore.
8. `POST /v1/ingest/git` still works.
9. Compile context for a matching task with a **verified architecture**
   statement plus the PR observation: architecture ranks first.
10. OIDC membership: a user without membership on that repository gets 403 on
    trigger and list.
11. MCP tool list still has 10 tools.
12. `go test ./...` and `go vet ./...` green.
