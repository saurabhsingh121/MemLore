package queries_test

import (
	"context"
	"testing"
	"time"

	appcontext "github.com/memlore/memlore/internal/application/context"
	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/application/queries"
	"github.com/memlore/memlore/internal/domain"
)

type stubSearcher struct {
	result queries.SearchKnowledgeResult
	err    error
	calls  []queries.SearchKnowledgeQuery
}

func (s *stubSearcher) Handle(_ context.Context, q queries.SearchKnowledgeQuery) (queries.SearchKnowledgeResult, error) {
	s.calls = append(s.calls, q)
	return s.result, s.err
}

func TestCompileContextRequiresTaskAndScope(t *testing.T) {
	handler := queries.NewCompileContextHandler(&stubSearcher{})
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")

	_, err := handler.Handle(context.Background(), queries.CompileContextQuery{})
	if err == nil {
		t.Fatal("expected validation error for task")
	}

	_, err = handler.Handle(context.Background(), queries.CompileContextQuery{Task: "do thing"})
	if err == nil {
		t.Fatal("expected validation error for scope")
	}

	_, err = handler.Handle(context.Background(), queries.CompileContextQuery{
		Task:  "do thing",
		Scope: scope,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCompileContextOmitsStaleAndSurfacesConflicts(t *testing.T) {
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	succ := "succ-id"
	stub := &stubSearcher{
		result: queries.SearchKnowledgeResult{
			Governance: []domain.LoreEntry{
				{
					ID: "a", Statement: "Use blue-green", Scope: scope,
					Origin: domain.KnowledgeOriginHumanAuthored, VerificationStatus: domain.VerificationUnverified,
					CreatedAt: now, UpdatedAt: now,
				},
				{
					ID: "b", Statement: "Use rolling", Scope: scope,
					Origin: domain.KnowledgeOriginHumanAuthored, VerificationStatus: domain.VerificationUnverified,
					CreatedAt: now, UpdatedAt: now,
				},
				{
					ID: "stale", Statement: "Old rule", Scope: scope,
					Origin: domain.KnowledgeOriginHumanAuthored, VerificationStatus: domain.VerificationVerified,
					SupersededByID: &succ, CreatedAt: now, UpdatedAt: now,
				},
				{
					ID: "inv", Statement: "Bad rule", Scope: scope,
					Origin: domain.KnowledgeOriginHumanAuthored, VerificationStatus: domain.VerificationInvalidated,
					CreatedAt: now, UpdatedAt: now,
				},
			},
			Warnings: []string{"graph_service_unavailable"},
		},
	}
	handler := queries.NewCompileContextHandler(stub)

	result, err := handler.Handle(context.Background(), queries.CompileContextQuery{
		Task:  "deploy",
		Scope: scope,
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if len(stub.calls) != 1 || stub.calls[0].IncludeStale {
		t.Fatalf("search calls = %+v", stub.calls)
	}
	for _, item := range result.Items {
		if item.ID == "stale" || item.ID == "inv" {
			t.Fatalf("stale item in packet: %+v", item)
		}
	}
	if len(result.Items) != 2 {
		t.Fatalf("items = %+v", result.Items)
	}
	if len(result.Conflicts) != 1 {
		t.Fatalf("conflicts = %+v", result.Conflicts)
	}
	if len(result.Conflicts[0].EntryIDs) != 2 {
		t.Fatalf("conflict ids = %v", result.Conflicts[0].EntryIDs)
	}
	if len(result.Warnings) != 1 || result.Warnings[0] != "graph_service_unavailable" {
		t.Fatalf("warnings = %v", result.Warnings)
	}
}

func TestCompileContextConflictSurvivesBudget(t *testing.T) {
	now := time.Now().UTC()
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	longA := stringsRepeat('a', 400)
	longB := stringsRepeat('b', 400)
	stub := &stubSearcher{
		result: queries.SearchKnowledgeResult{
			Governance: []domain.LoreEntry{
				{ID: "a", Statement: longA, Scope: scope, CreatedAt: now, UpdatedAt: now},
				{ID: "b", Statement: longB, Scope: scope, CreatedAt: now, UpdatedAt: now},
			},
			Warnings: []string{},
		},
	}
	handler := queries.NewCompileContextHandler(stub)

	result, err := handler.Handle(context.Background(), queries.CompileContextQuery{
		Task:        "task",
		Scope:       scope,
		TokenBudget: 200,
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("items = %d", len(result.Items))
	}
	if len(result.Conflicts) != 1 || len(result.Conflicts[0].EntryIDs) != 2 {
		t.Fatalf("conflicts = %+v", result.Conflicts)
	}
}

func TestCompileContextRanksAndBudgets(t *testing.T) {
	now := time.Date(2026, 8, 27, 12, 0, 0, 0, time.UTC)
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")
	stub := &stubSearcher{
		result: queries.SearchKnowledgeResult{
			Query: "payment outbox",
			Scope: &scope,
			Governance: []domain.LoreEntry{{
				ID:                 "gov-1",
				Statement:          "Verified outbox rule",
				Scope:              scope,
				Origin:             domain.KnowledgeOriginHumanAuthored,
				VerificationStatus: domain.VerificationVerified,
				CreatedAt:          now,
				UpdatedAt:          now,
			}},
			Graph: []ports.GraphFact{{
				ID:        "fact-1",
				Statement: "Graph outbox hint",
				Score:     0.95,
			}},
			Warnings: []string{},
		},
	}
	handler := queries.NewCompileContextHandler(stub)

	result, err := handler.Handle(context.Background(), queries.CompileContextQuery{
		Task:        "Implement outbox",
		Query:       "payment outbox",
		Scope:       scope,
		TokenBudget: 4096,
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if len(result.Items) != 2 {
		t.Fatalf("items = %+v", result.Items)
	}
	if result.Items[0].Source != appcontext.ItemSourceGovernance {
		t.Fatalf("first item source = %s", result.Items[0].Source)
	}
	if result.Meta.ItemsTotalRanked != 2 {
		t.Fatalf("meta = %+v", result.Meta)
	}
}

func TestCompileContextDefaultsQueryToTask(t *testing.T) {
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	stub := &stubSearcher{result: queries.SearchKnowledgeResult{Warnings: []string{}}}
	handler := queries.NewCompileContextHandler(stub)

	_, err := handler.Handle(context.Background(), queries.CompileContextQuery{
		Task:  "my task text",
		Scope: scope,
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if len(stub.calls) != 1 || stub.calls[0].Query != "my task text" {
		t.Fatalf("calls = %+v", stub.calls)
	}
}

func TestCompileContextTokenBudgetLimitsItems(t *testing.T) {
	now := time.Now().UTC()
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	longStatement := stringsRepeat('x', 400)

	stub := &stubSearcher{
		result: queries.SearchKnowledgeResult{
			Governance: []domain.LoreEntry{
				{ID: "1", Statement: longStatement, Scope: scope, CreatedAt: now, UpdatedAt: now},
				{ID: "2", Statement: longStatement + " extra", Scope: scope, CreatedAt: now, UpdatedAt: now},
			},
			Warnings: []string{},
		},
	}
	handler := queries.NewCompileContextHandler(stub)

	result, err := handler.Handle(context.Background(), queries.CompileContextQuery{
		Task:        "task",
		Scope:       scope,
		TokenBudget: 200,
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("items = %d want 1", len(result.Items))
	}
}

func stringsRepeat(ch byte, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = ch
	}
	return string(b)
}
