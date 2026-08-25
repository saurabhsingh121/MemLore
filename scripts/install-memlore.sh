#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

mkdir -p bin
go build -o bin/memlore ./cmd/memlore

echo "Built: $ROOT/bin/memlore"
echo ""
echo "Cross-project MCP (Cursor .cursor/mcp.json):"
echo '  "command": "'"$ROOT"'/bin/memlore",'
echo '  "args": ["mcp"]'
echo ""
echo "Optional: add bin/ to PATH or symlink:"
echo "  ln -sf $ROOT/bin/memlore /usr/local/bin/memlore"
