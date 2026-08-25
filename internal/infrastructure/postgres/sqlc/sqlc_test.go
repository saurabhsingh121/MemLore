package sqlc_test

import (
	"testing"

	"github.com/memlore/memlore/internal/infrastructure/postgres/sqlc"
)

func TestSQLcPackageExportsGeneratedTypes(t *testing.T) {
	_ = sqlc.New(nil)
	_ = sqlc.LoreEntry{}
	_ = sqlc.AuditRecord{}
	_ = sqlc.InsertLoreEntryParams{}
	_ = sqlc.InsertAuditRecordParams{}
}
