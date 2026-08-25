package commands_test

import (
	"context"
	"testing"
	"time"

	"github.com/memlore/memlore/internal/application/commands"
	"github.com/memlore/memlore/internal/domain"
	"github.com/memlore/memlore/internal/infrastructure/clock"
	"github.com/memlore/memlore/internal/infrastructure/memory"
)

func TestCreateLorePersistsEntryAndCreateAudit(t *testing.T) {
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	now := time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)
	handler := commands.NewCreateLoreHandler(begin, clock.FixedClock{Instant: now})
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")

	entry, err := handler.Handle(context.Background(), commands.CreateLoreCommand{
		Statement: "Rule",
		Scope:     scope,
		ActorID:   "alice",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	audits, err := uow.Audits().ListByTarget(context.Background(), entry.ID)
	if err != nil {
		t.Fatalf("ListByTarget: %v", err)
	}
	if len(audits) != 1 || audits[0].Action != domain.AuditActionCreate {
		t.Fatalf("audits = %+v", audits)
	}
}
