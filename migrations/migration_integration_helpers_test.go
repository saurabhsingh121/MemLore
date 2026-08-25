//go:build integration

package migrations_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func extractGooseUpSQL(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	body, err := os.ReadFile(filepath.Join(filepath.Dir(file), "00001_lore_audit.sql"))
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	content := string(body)
	const upStart = "-- +goose Up"
	const upEnd = "-- +goose Down"
	start := strings.Index(content, upStart)
	if start < 0 {
		t.Fatal("missing goose Up marker")
	}
	end := strings.Index(content, upEnd)
	if end < 0 {
		t.Fatal("missing goose Down marker")
	}
	up := content[start:end]
	up = strings.ReplaceAll(up, "-- +goose Up", "")
	up = strings.ReplaceAll(up, "-- +goose StatementBegin", "")
	up = strings.ReplaceAll(up, "-- +goose StatementEnd", "")
	return strings.TrimSpace(up)
}

func replaceDatabaseInDSN(dsn, dbName string) string {
	if idx := strings.LastIndex(dsn, "/"); idx >= 0 {
		return dsn[:idx+1] + dbName
	}
	return dsn + "/" + dbName
}
