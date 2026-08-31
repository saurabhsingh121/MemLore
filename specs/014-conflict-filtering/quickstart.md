# Quickstart: F112 temporal filtering + conflicts

## Local verify

```bash
go test ./internal/application/context/... ./internal/application/queries/... ./internal/adapters/...
go test ./...
```

## Dogfood scenario

1. Create two current lore entries in the same scope with different statements
   (contradictory rules).
2. Supersede an older third entry (or invalidate one) in that scope.
3. Call `memlore.get_for_task` / `POST /v1/context/compile` for that scope.
4. Confirm:
   - `items` omit the superseded/invalidated entry
   - `conflicts` lists both current disagreeing ids
   - `memlore.get` on the stale id still returns it
5. Call list/search with `include_stale=true` and confirm stale appears.

## Expected packet fragment

```json
{
  "conflicts": [
    {
      "scope": { "kind": "repository", "key": "demo" },
      "entry_ids": ["...", "..."],
      "statements": ["Rule A", "Rule B"]
    }
  ]
}
```
