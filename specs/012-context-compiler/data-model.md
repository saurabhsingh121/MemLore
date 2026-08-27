# Data Model: Context Compiler (F109)

## CompileContextQuery

| Field | Type | Required |
|-------|------|----------|
| task | string | yes |
| query | string | no (defaults to task) |
| scope | Scope | yes |
| token_budget | int | no (default 4096) |

## ContextItem

| Field | Type | Notes |
|-------|------|-------|
| id | string | Lore id or graph fact id |
| statement | string | Natural language |
| source | enum | `governance` \| `graph` |
| authority_score | number | 0–1 ranking score |
| authority_factors | object | Explainable factors |
| scope | Scope | |
| evidence | Evidence[] | governance only |
| provenance_refs | string[] | graph only |

## ContextPacket (response)

| Field | Type |
|-------|------|
| task | string |
| query | string |
| scope | Scope |
| items | ContextItem[] |
| meta | ContextMeta |
| warnings | string[] |

## ContextMeta

| Field | Type |
|-------|------|
| token_budget | int |
| estimated_tokens | int |
| items_included | int |
| items_total_ranked | int |
