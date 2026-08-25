package layout_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestGoModuleLayoutContract(t *testing.T) {
	root := repoRoot(t)
	required := []string{
		"go.mod",
		"go.sum",
		"sqlc.yaml",
		"cmd/memlore/main.go",
		"migrations/00001_lore_audit.sql",
		"db/queries/smoke.sql",
		"internal/domain/lore.go",
		"internal/application/ports/repositories.go",
		"internal/adapters/doc.go",
		"internal/infrastructure/postgres/doc.go",
		"internal/infrastructure/postgres/sqlc/db.go",
	}
	for _, rel := range required {
		path := filepath.Join(root, rel)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("required path missing: %s (%v)", rel, err)
		}
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// internal/layout -> repo root
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
