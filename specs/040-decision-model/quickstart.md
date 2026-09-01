# Quickstart: First-Class Decision Model (F040)

1. Seed repository `github.com/acme/payments` with:
   - one F032 accepted-ADR lore row (verified `architecture_decision` + `adr`)
   - one unverified git observational lore row
   - one pending F035-eligible PR observation (not accepted)
2. `memlore decision create --repository github.com/acme/payments --question "How should payment events be published?" --choice "Transactional outbox" --owner alice --alternative "Dual-write" --actor alice`
   — stores a current human Decision; get returns question, choice, owner,
   alternative, `source: human`.
3. `memlore decision list --repository github.com/acme/payments` — exactly
   two current items (human + ADR projection); git and PR observations
   absent; ADR appears once with `source: adr`.
4. `GET /v1/decisions?scope_kind=repository&scope_key=github.com/acme/payments`
   matches the CLI list. `GET /v1/decisions/{adr-lore-id}` returns the
   projected ADR Decision.
5. `memlore decision supersede <human-id> --question "…" --choice "Outbox + idempotent consumers" --owner alice --actor alice`
   — successor current; predecessor gettable and absent from list-current.
6. `memlore decision supersede <adr-lore-id> --question "…" --choice "…" --owner alice --actor alice`
   — human successor current; ADR lore superseded not deleted; list-current
   does not show the old ADR choice as a second current fact.
7. Compile / `get_for_task` for a matching task: current Decisions appear
   under section `decisions`; they outrank leftover unverified observations;
   remaining ADR (if any current) still outranks leftover observations.
8. `POST /v1/review-queue/{pr-id}/accept` — still lore only; decision list
   unchanged.
9. `POST /v1/lore-entries` — still unverified `human_authored` lore, not a
   Decision.
10. OIDC membership: a user without membership gets 403 on list/create; get
    of a foreign id is 404; a reader cannot create/supersede.
11. MCP tool list still has 10 tools.
12. `go test ./...` and `go vet ./...` green.
