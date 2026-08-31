# Data Model: Repository Intelligence Profile (F020)

No new persistence. On-read types only.

## ProfileSectionID

Stable ids (first-match classification order, most specific first):

1. `decisions`
2. `conventions`
3. `gotchas`
4. `migrations`
5. `ownership`
6. `operational_risks`
7. `hotspots`
8. `related_services`
9. `architecture`
10. `technologies`
11. `recent_changes`

## Classification cues (deterministic)

| Section | Cues (case-insensitive substring unless noted) |
|---------|------------------------------------------------|
| decisions | evidence type `adr`; origin `architecture_decision`; `adr-`, `adr `, `decision`, `we chose`, `instead of` |
| conventions | `convention`, `must not`, `must never`, `coding standard` |
| gotchas | `gotcha`, `pitfall`, `caveat`, `watch out`, `never `, `eventually consistent` |
| migrations | `migrat` (migration/migrating) |
| ownership | `owner`, `owned by`, `codeowners` |
| operational_risks | `outage`, `incident`, `on-call`, `operational risk` |
| hotspots | `hotspot`, `fragile`, `frequently changed` |
| related_services | `depends on`, `related service`, `publishes`, `consumes`, `dependency` |
| architecture | `architecture`, `hexagonal`, `layered`, `microservice` |
| technologies | `postgres`, `postgresql`, `kafka`, `redis`, `neo4j`, `java `, `golang`, `python ` |
| recent_changes | `recent change`, `changelog`, `as of 20` |

First match in table order wins.

## RepositoryProfile (response)

| Field | Meaning |
|-------|---------|
| repository | `{ kind, key }` — always `repository` |
| sections | map/list of `{ id, items[] }` — only non-empty |
| conflicts | same shape as compile |
| warnings | e.g. `graph_service_unavailable` |
| meta | token_budget, estimated_tokens, items_included, items_total_ranked, unclassified_count |

## ProfileItem

Same as compiled context item fields: id, statement, source, authority_score,
trust_band, authority_factors, scope, evidence, provenance_refs.
`section` is implied by parent.

## Validation

- `scope.kind` MUST be `repository`
- `scope.key` non-empty (existing scope rules)
- `token_budget` ≤ 0 means default 4096
