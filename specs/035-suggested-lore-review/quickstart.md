# Quickstart: Suggested Lore Review Queue (F035)

1. Seed repository `github.com/acme/payments` with:
   - one git observational lore row (commit evidence, unverified)
   - one PR observational lore row (pr evidence, unverified)
   - one F032 accepted-ADR lore row (verified `architecture_decision`)
2. `memlore review list --repository github.com/acme/payments` — exactly two
   pending items; ADR absent; no invented confidence/reason.
3. `GET /v1/review-queue?scope_kind=repository&scope_key=github.com/acme/payments`
   (local `X-Memlore-Actor` not required for GET) matches the CLI list.
4. `memlore review accept <git-id> --actor alice` — new current
   `human_verified` + verified lore with the same commit evidence;
   observational predecessor is superseded, not deleted.
5. `POST /v1/review-queue/<pr-id>/accept` with a different `statement` —
   successor origin `human_authored`, statement is the edit, PR evidence kept.
6. Seed a third observational extract; `memlore review reject <id> --actor alice`
   — pending list no longer includes it; lore row still observational.
7. Re-run git ingest for a rejected SHA — still not pending.
8. `POST /v1/lore-entries/{obs-id}/verify` on a leftover observation — origin
   still `repository_observation` (not Accept).
9. Compile a matching task with F032 ADR + F035-accepted item + leftover
   unverified observation: accepted item outranks the leftover observation;
   ADR still outranks leftover observations.
10. OIDC membership: a user without membership on that repository gets 403 on
    list and accept. A reader cannot accept.
11. MCP tool list still has 10 tools.
12. `GET /v1/ingest/candidates` still lists observational/ADR-evidence lore and
    is not the review workflow.
13. `go test ./...` and `go vet ./...` green.
