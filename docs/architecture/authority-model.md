# Authority model

Authority in MemLore is **explainable**. It is derived from explicit factors,
not a single opaque magic score. v1 (F003) evaluates factors **ephemerally**
at compile and explain time from existing lore/graph fields. Scores are never
stored without their factors; v1 stores neither.

## Factors (v1)

| Factor | Role | v1 source |
|--------|------|-----------|
| source_type | ADR, human statement, agent observation/inference, repo observation, import, graph | derived from origin + evidence, or `graph` for graph-plane hits |
| origin | who produced the knowledge | LoreEntry.Origin |
| verification_status | unverified, verified, invalidated | LoreEntry |
| evidence_strength | ADR `1.0`, any other evidence `0.6`, none `0.0` | Evidence[] |
| recency | 0.00–0.10 boost decaying linearly to zero at 365 days | CreatedAt |
| scope_match | exact `1.0`, same kind `0.5`, none `0.0` | entry vs requested compile scope |
| supersession_status | current vs superseded | `superseded_by_id` |
| graph_score | raw graph retrieval score | GraphFact.Score (graph plane only) |
| source_reliability | historical reliability of the source | **deferred** (omitted in v1) |

A ranking score is computed from these factors, but the factors remain on
ContextPacket items and explain payloads so every important retrieval can
answer:

- Why was this returned?
- Who said it?
- Where did it come from?
- What evidence supports it?
- How authoritative is it?
- Is it current?
- Has something superseded it?

## Trust bands (v1)

Assigned from factors (not from score cut-points):

| Band | Meaning |
|------|---------|
| `canonical` | Current, verified, ADR-class, human-side origin |
| `high` | Current verified with weaker/no evidence; or **verified** agent origin |
| `medium` | Current unverified human-side; superseded (capped) |
| `low` | Graph-only; unverified agent origin; unverified repo observation |
| `untrusted` | Invalidated |

## Hard rules

- Agent inference MUST NOT silently gain human authority.
  Unverified `agent_inference` / `agent_observation` cannot be `canonical` or
  `high`. Verified agent origin can be `high` but never `canonical`.
- Repository evidence is strong observational evidence, not automatic intent.
- Conflicts are preserved and surfaced.
- Invalidated scores are capped at `0.20` if they reach the scorer.
- Ranking order intent: verified+strong evidence ≫ verified weaker ≫
  unverified human ≫ graph ≫ unverified agent inference.

See `specs/016-authority-factors/` for the full evaluation contract.
