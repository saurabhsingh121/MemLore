# Quickstart: F003 authority factors

## Local verify

```bash
go test ./internal/domain/ ./internal/application/authority/ ./internal/application/context/ ./internal/application/queries/ ./internal/adapters/...
go test ./...
```

## Dogfood scenario

1. Create lore with ADR evidence; verify it.
2. Create unverified human lore in the same scope.
3. Compile context / `memlore.get_for_task`.
4. Confirm:
   - verified ADR item has `trust_band` `canonical` or `high` and
     `evidence_strength` / `source_type` in `authority_factors`
   - unverified item is `medium` and ranks below
5. `memlore.explain` the verified id: `trust_band`, `authority_factors`,
   `factor_breakdown` present; no `summary`.
6. REST: `GET /v1/lore-entries/{id}/explain` matches MCP.

## Expected item fragment

```json
{
  "trust_band": "canonical",
  "authority_factors": {
    "verification_status": "verified",
    "origin": "human_authored",
    "source_type": "adr",
    "evidence_strength": 1.0,
    "supersession_status": "current"
  }
}
```
