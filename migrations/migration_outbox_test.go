package migrations_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestOutboxMigrationDefinesExpectedTableAndIndexes(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	sqlPath := filepath.Join(filepath.Dir(file), "00002_outbox_events.sql")
	body, err := os.ReadFile(sqlPath)
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	content := string(body)

	required := []string{
		"CREATE TABLE outbox_events",
		"ux_outbox_events_idempotency_key",
		"ix_outbox_events_status_created",
		"payload JSONB",
	}
	for _, fragment := range required {
		if !strings.Contains(content, fragment) {
			t.Errorf("migration missing %q", fragment)
		}
	}
}
