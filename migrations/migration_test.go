package migrations_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestLoreAuditMigrationDefinesExpectedTablesAndIndexes(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	sqlPath := filepath.Join(filepath.Dir(file), "00001_lore_audit.sql")
	body, err := os.ReadFile(sqlPath)
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	content := string(body)

	required := []string{
		"CREATE TABLE lore_entries",
		"CREATE TABLE audit_records",
		"CREATE INDEX ix_lore_entries_scope_created",
		"CREATE INDEX ix_audit_records_target_id",
		"CREATE INDEX ix_audit_records_target_created",
		"scope_kind VARCHAR(64)",
		"evidence JSONB",
	}
	for _, fragment := range required {
		if !strings.Contains(content, fragment) {
			t.Errorf("migration missing %q", fragment)
		}
	}
}
