package queries_test

import (
	"context"
	"strings"
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
	handler := queries.NewCompileContextHandler(&stubSearcher{}, nil)
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
			Governance: queries.HitsFromEntries([]domain.LoreEntry{
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
			}),
			Warnings: []string{"graph_service_unavailable"},
		},
	}
	handler := queries.NewCompileContextHandler(stub, nil)

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
			Governance: queries.HitsFromEntries([]domain.LoreEntry{
				{ID: "a", Statement: longA, Scope: scope, CreatedAt: now, UpdatedAt: now},
				{ID: "b", Statement: longB, Scope: scope, CreatedAt: now, UpdatedAt: now},
			}),
			Warnings: []string{},
		},
	}
	handler := queries.NewCompileContextHandler(stub, nil)

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
			Governance: queries.HitsFromEntries([]domain.LoreEntry{{
				ID:                 "gov-1",
				Statement:          "Verified outbox rule",
				Scope:              scope,
				Origin:             domain.KnowledgeOriginHumanAuthored,
				VerificationStatus: domain.VerificationVerified,
				CreatedAt:          now,
				UpdatedAt:          now,
			}}),
			Graph: []ports.GraphFact{{
				ID:        "fact-1",
				Statement: "Graph outbox hint",
				Score:     0.95,
			}},
			Warnings: []string{},
		},
	}
	handler := queries.NewCompileContextHandler(stub, nil)

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
	if result.Items[0].TrustBand == "" {
		t.Fatal("missing trust_band on compiled item")
	}
	if result.Items[0].AuthorityFactors.SourceType == "" {
		t.Fatal("missing source_type on compiled item")
	}
	if result.Meta.ItemsTotalRanked != 2 {
		t.Fatalf("meta = %+v", result.Meta)
	}
}

func TestCompileContextDefaultsQueryToTask(t *testing.T) {
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	stub := &stubSearcher{result: queries.SearchKnowledgeResult{Warnings: []string{}}}
	handler := queries.NewCompileContextHandler(stub, nil)

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

func TestCompileContextPacketSectionsAndOmitsEmpty(t *testing.T) {
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")
	stub := &stubSearcher{
		result: queries.SearchKnowledgeResult{
			Governance: queries.HitsFromEntries([]domain.LoreEntry{
				{
					ID: "arch", Statement: "Hexagonal architecture with ports.", Scope: scope,
					Origin: domain.KnowledgeOriginHumanAuthored, VerificationStatus: domain.VerificationVerified,
					CreatedAt: now, UpdatedAt: now,
				},
				{
					ID: "dec", Statement: "Use Kafka instead of RabbitMQ.", Scope: scope,
					Origin: domain.KnowledgeOriginHumanAuthored, VerificationStatus: domain.VerificationVerified,
					Evidence:  []domain.EvidenceReference{{Type: domain.EvidenceTypeADR, Value: "ADR-017"}},
					CreatedAt: now, UpdatedAt: now,
				},
				{
					ID: "task", Statement: "Payment outbox must persist events atomically.", Scope: scope,
					Origin: domain.KnowledgeOriginHumanAuthored, VerificationStatus: domain.VerificationUnverified,
					CreatedAt: now, UpdatedAt: now,
				},
				{
					ID: "noise", Statement: "The sky is blue.", Scope: scope,
					Origin: domain.KnowledgeOriginHumanAuthored, VerificationStatus: domain.VerificationUnverified,
					CreatedAt: now, UpdatedAt: now,
				},
			}),
			Warnings: []string{},
		},
	}
	handler := queries.NewCompileContextHandler(stub, nil)
	result, err := handler.Handle(context.Background(), queries.CompileContextQuery{
		Task:  "Implement payment outbox",
		Scope: scope,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Items) != 4 {
		t.Fatalf("items = %d, want 4", len(result.Items))
	}
	got := map[appcontext.ProfileSectionID]int{}
	for _, sec := range result.Sections {
		if len(sec.Items) == 0 {
			t.Fatalf("empty section %s", sec.ID)
		}
		got[sec.ID] = len(sec.Items)
	}
	if got[appcontext.SectionArchitecture] != 1 || got[appcontext.SectionDecisions] != 1 {
		t.Fatalf("sections = %v", got)
	}
	if got[appcontext.SectionTaskContext] != 1 {
		t.Fatalf("task_context missing: %v", got)
	}
	if _, ok := got[appcontext.SectionConventions]; ok {
		t.Fatal("empty conventions must be omitted")
	}
	if result.Meta.UnclassifiedCount != 1 {
		t.Fatalf("unclassified = %d", result.Meta.UnclassifiedCount)
	}
	if len(result.Sources) != 1 || result.Sources[0].Value != "ADR-017" {
		t.Fatalf("sources = %+v", result.Sources)
	}
}

func TestCompileContextMergesRepositoryBriefing(t *testing.T) {
	now := time.Now().UTC()
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")
	list := &stubLister{entries: []domain.LoreEntry{
		{
			ID: "arch", Statement: "Hexagonal architecture with ports.", Scope: scope,
			Origin: domain.KnowledgeOriginHumanAuthored, VerificationStatus: domain.VerificationVerified,
			CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "noise", Statement: "The sky is blue.", Scope: scope,
			Origin: domain.KnowledgeOriginHumanAuthored, CreatedAt: now, UpdatedAt: now,
		},
	}}
	stub := &stubSearcher{
		result: queries.SearchKnowledgeResult{
			Governance: queries.HitsFromEntries([]domain.LoreEntry{{
				ID: "task", Statement: "Payment outbox must persist events atomically.", Scope: scope,
				Origin: domain.KnowledgeOriginHumanAuthored, CreatedAt: now, UpdatedAt: now,
			}}),
			Warnings: []string{},
		},
	}
	handler := queries.NewCompileContextHandler(stub, list)
	result, err := handler.Handle(context.Background(), queries.CompileContextQuery{
		Task:  "Implement payment outbox",
		Scope: scope,
	})
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]bool{}
	for _, item := range result.Items {
		ids[item.ID] = true
	}
	if !ids["arch"] || !ids["task"] {
		t.Fatalf("items = %v", ids)
	}
	if ids["noise"] {
		t.Fatal("unclassified list lore must not be merged")
	}
}

func TestCompileContextFilesInfluenceTaskContext(t *testing.T) {
	now := time.Now().UTC()
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	stub := &stubSearcher{
		result: queries.SearchKnowledgeResult{
			Governance: queries.HitsFromEntries([]domain.LoreEntry{
				{ID: "file", Statement: "Publisher writes to src/payments/outbox.go after commit.", Scope: scope, CreatedAt: now, UpdatedAt: now},
				{ID: "other", Statement: "Unrelated refund SLA is two days.", Scope: scope, CreatedAt: now, UpdatedAt: now},
			}),
			Warnings: []string{},
		},
	}
	handler := queries.NewCompileContextHandler(stub, nil)
	result, err := handler.Handle(context.Background(), queries.CompileContextQuery{
		Task:         "wire publisher",
		Scope:        scope,
		ChangedFiles: []string{"src/payments/outbox.go"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(stub.calls) != 1 || !strings.Contains(stub.calls[0].Query, "src/payments/outbox.go") {
		t.Fatalf("search query = %+v", stub.calls)
	}
	if len(result.Sections) != 1 || result.Sections[0].ID != appcontext.SectionTaskContext {
		t.Fatalf("sections = %+v", result.Sections)
	}
	if result.Sections[0].Items[0].ID != "file" {
		t.Fatalf("task_context = %+v", result.Sections[0].Items)
	}
}

func TestCompileContextAgentIDDoesNotChangeRanking(t *testing.T) {
	now := time.Now().UTC()
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	stub := &stubSearcher{
		result: queries.SearchKnowledgeResult{
			Governance: queries.HitsFromEntries([]domain.LoreEntry{
				{ID: "a", Statement: "Hexagonal architecture.", Scope: scope, Origin: domain.KnowledgeOriginHumanAuthored, VerificationStatus: domain.VerificationVerified, CreatedAt: now, UpdatedAt: now},
				{ID: "b", Statement: "Maybe use hexagonal architecture.", Scope: scope, Origin: domain.KnowledgeOriginAgentInference, VerificationStatus: domain.VerificationUnverified, CreatedAt: now, UpdatedAt: now},
			}),
			Warnings: []string{},
		},
	}
	handler := queries.NewCompileContextHandler(stub, nil)
	base := queries.CompileContextQuery{Task: "review architecture", Scope: scope}
	withAgent := base
	withAgent.AgentID = "cursor-agent"
	a, err := handler.Handle(context.Background(), base)
	if err != nil {
		t.Fatal(err)
	}
	b, err := handler.Handle(context.Background(), withAgent)
	if err != nil {
		t.Fatal(err)
	}
	if b.AgentID != "cursor-agent" {
		t.Fatalf("agent_id = %q", b.AgentID)
	}
	if len(a.Items) != len(b.Items) {
		t.Fatalf("item count %d vs %d", len(a.Items), len(b.Items))
	}
	for i := range a.Items {
		if a.Items[i].ID != b.Items[i].ID {
			t.Fatalf("ranking changed: %+v vs %+v", a.Items[i], b.Items[i])
		}
		if diff := a.Items[i].AuthorityScore - b.Items[i].AuthorityScore; diff > 1e-9 || diff < -1e-9 {
			t.Fatalf("score changed: %v vs %v", a.Items[i].AuthorityScore, b.Items[i].AuthorityScore)
		}
		if b.Items[i].AuthorityFactors.Origin == "cursor-agent" {
			t.Fatal("agent_id must not appear in authority factors")
		}
	}
	if a.Items[0].ID != "a" {
		t.Fatalf("verified architecture should rank first: %+v", a.Items)
	}
}

func TestCompileContextSearchIncludesTicket(t *testing.T) {
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	stub := &stubSearcher{result: queries.SearchKnowledgeResult{Warnings: []string{}}}
	handler := queries.NewCompileContextHandler(stub, nil)
	_, err := handler.Handle(context.Background(), queries.CompileContextQuery{
		Task:   "do thing",
		Query:  "outbox",
		Ticket: "PAY-1842",
		Scope:  scope,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(stub.calls) != 1 || stub.calls[0].Query != "outbox PAY-1842" {
		t.Fatalf("calls = %+v", stub.calls)
	}
}

func TestCompileContextTokenBudgetLimitsItems(t *testing.T) {
	now := time.Now().UTC()
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	longStatement := stringsRepeat('x', 400)

	stub := &stubSearcher{
		result: queries.SearchKnowledgeResult{
			Governance: queries.HitsFromEntries([]domain.LoreEntry{
				{ID: "1", Statement: longStatement, Scope: scope, CreatedAt: now, UpdatedAt: now},
				{ID: "2", Statement: longStatement + " extra", Scope: scope, CreatedAt: now, UpdatedAt: now},
			}),
			Warnings: []string{},
		},
	}
	handler := queries.NewCompileContextHandler(stub, nil)

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

func TestCompileContextVerifiedArchitectureOutranksGitObservation(t *testing.T) {
	now := time.Now().UTC()
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")
	gitEv := []domain.EvidenceReference{{Type: domain.EvidenceTypeCommit, Value: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}}
	stub := &stubSearcher{
		result: queries.SearchKnowledgeResult{
			Governance: queries.HitsFromEntries([]domain.LoreEntry{
				{ID: "git", Statement: "Use the outbox because dual-writes race.", Scope: scope, Origin: domain.KnowledgeOriginRepositoryObservation, VerificationStatus: domain.VerificationUnverified, Evidence: gitEv, CreatedAt: now, UpdatedAt: now},
				{ID: "arch", Statement: "Hexagonal architecture is canonical.", Scope: scope, Origin: domain.KnowledgeOriginHumanAuthored, VerificationStatus: domain.VerificationVerified, CreatedAt: now, UpdatedAt: now},
			}),
			Warnings: []string{},
		},
	}
	handler := queries.NewCompileContextHandler(stub, nil)
	result, err := handler.Handle(context.Background(), queries.CompileContextQuery{
		Task:  "review architecture outbox",
		Scope: scope,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Items) < 2 {
		t.Fatalf("items = %+v", result.Items)
	}
	if result.Items[0].ID != "arch" {
		t.Fatalf("expected verified architecture first, got %+v", result.Items)
	}
	var gitItem appcontext.RankedItem
	for _, it := range result.Items {
		if it.ID == "git" {
			gitItem = it
		}
	}
	if gitItem.ID == "" {
		t.Fatal("git observation missing from packet")
	}
	if gitItem.AuthorityFactors.Origin != string(domain.KnowledgeOriginRepositoryObservation) {
		t.Fatalf("git origin = %+v", gitItem.AuthorityFactors)
	}
}

func TestCompileContextVerifiedArchitectureOutranksPRObservation(t *testing.T) {
	now := time.Now().UTC()
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")
	prEv := []domain.EvidenceReference{{Type: domain.EvidenceTypePR, Value: "acme/payments#1842"}}
	stub := &stubSearcher{
		result: queries.SearchKnowledgeResult{
			Governance: queries.HitsFromEntries([]domain.LoreEntry{
				{ID: "pr", Statement: "Use the outbox because dual-writes race.", Scope: scope, Origin: domain.KnowledgeOriginRepositoryObservation, VerificationStatus: domain.VerificationUnverified, Evidence: prEv, CreatedAt: now, UpdatedAt: now},
				{ID: "arch", Statement: "Hexagonal architecture is canonical.", Scope: scope, Origin: domain.KnowledgeOriginHumanAuthored, VerificationStatus: domain.VerificationVerified, CreatedAt: now, UpdatedAt: now},
			}),
			Warnings: []string{},
		},
	}
	handler := queries.NewCompileContextHandler(stub, nil)
	result, err := handler.Handle(context.Background(), queries.CompileContextQuery{
		Task:  "review architecture outbox",
		Scope: scope,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Items) < 2 {
		t.Fatalf("items = %+v", result.Items)
	}
	if result.Items[0].ID != "arch" {
		t.Fatalf("expected verified architecture first, got %+v", result.Items)
	}
	var prItem appcontext.RankedItem
	for _, it := range result.Items {
		if it.ID == "pr" {
			prItem = it
		}
	}
	if prItem.ID == "" {
		t.Fatal("PR observation missing from packet")
	}
	if prItem.AuthorityFactors.Origin != string(domain.KnowledgeOriginRepositoryObservation) {
		t.Fatalf("pr origin = %+v", prItem.AuthorityFactors)
	}
}

func stringsRepeat(ch byte, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = ch
	}
	return string(b)
}
