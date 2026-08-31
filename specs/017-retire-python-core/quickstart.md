# Quickstart: retire Python core

```bash
# After this change, from repo root:
test ! -d src/memlore
test ! -d tests
test ! -f pyproject.toml
./scripts/install-memlore.sh
./bin/memlore version
go test ./...
# graph-service still:
cd graph-service && uv sync && uv run pytest -m "not integration"
```
