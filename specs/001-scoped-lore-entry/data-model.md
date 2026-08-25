# Data Model: Scoped Human-Authored Lore Entry

## Enumerations

### ScopeKind

| Value | Required in MVP |
|-------|-----------------|
| `team` | yes |
| `repository` | yes |
| `organization` | optional accept |
| `project` | optional accept |
| `feature` | optional accept |
| `task` | optional accept |

Unsupported / blank → validation error.

### EvidenceType

| Value | Required in MVP |
|-------|-----------------|
| `url` | yes |
| `path` | yes |
| `adr` | yes |

### KnowledgeOrigin

| Value | Notes |
|-------|-------|
| `human_authored` | only creatable origin this slice |
| `human_verified` | reserved / unused as create origin |
| `agent_observation` | reserved |
| `agent_inference` | reserved |
| `repository_observation` | reserved |
| `imported_source` | reserved |
| `architecture_decision` | reserved |

### VerificationStatus

| Value | Meaning |
|-------|---------|
| `unverified` | default on create |
| `verified` | after successful verify |

### AuditAction

| Value | When written |
|-------|----------------|
| `create` | on successful create |
| `verify` | on first successful transition to verified only |

## Entities

### LoreEntry

| Field | Type | Rules |
|-------|------|-------|
| id | UUID string | generated, immutable |
| statement | string | trimmed; non-empty; max 8000 chars |
| scope_kind | ScopeKind | required |
| scope_key | string | trimmed; non-empty; max 512 chars |
| origin | KnowledgeOrigin | always `human_authored` on create |
| verification_status | VerificationStatus | default `unverified` |
| evidence | list[EvidenceReference] | 0..N; each valid type+value |
| created_by | string | actor id; non-empty |
| created_at | datetime (UTC) | set on create |
| verified_by | string \| null | set on first verify |
| verified_at | datetime (UTC) \| null | set on first verify |
| updated_at | datetime (UTC) | changes on verify |

**Uniqueness**: Primary key on `id` only. No unique constraint on
`(scope_kind, scope_key, statement)`.

**Relationships**: Has many AuditRecords (by `target_id`). Evidence embedded
(JSONB) for MVP.

### EvidenceReference

| Field | Type | Rules |
|-------|------|-------|
| type | EvidenceType | required |
| value | string | trimmed; non-empty; max 2048 chars |

### AuditRecord

| Field | Type | Rules |
|-------|------|-------|
| id | UUID string | generated |
| target_id | UUID string | lore entry id; FK logical |
| action | AuditAction | `create` \| `verify` |
| actor_id | string | non-empty |
| created_at | datetime (UTC) | append-only |

**Ordering**: List by `created_at` ascending, then `id` ascending for
stability.

## State transitions

```text
[create] --> unverified
unverified --verify--> verified
verified --verify--> verified  (no-op; no new audit)
```

- Statement, origin, scope, evidence, created_by/created_at NEVER change on
  verify.
- No delete/invalidate/supersede transitions in this feature.

## Validation summary

- Reject blank actor on create/verify.
- Reject invalid scope kind/key, evidence type/value, oversized statement.
- Unknown lore id on get/verify/list-audits → not found.
- Create and first verify MUST occur in one DB transaction each with their
  audit insert (entry + audit atomic).

## Indexes (planned)

- `lore_entries (scope_kind, scope_key, created_at DESC)` for list-by-scope
- `audit_records (target_id, created_at ASC, id ASC)` for audit list
