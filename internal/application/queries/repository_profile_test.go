package queries_test

import (
	"context"
	"testing"
	"time"

	appcontext "github.com/memlore/memlore/internal/application/context"
	"github.com/memlore/memlore/internal/application/queries"
	"github.com/memlore/memlore/internal/domain"
)

type stubLister struct {
	entries []domain.LoreEntry
	err     error
}

func (s *stubLister) Handle(_ context.Context, _ queries.ListLoreByScopeQuery) ([]domain.LoreEntry, error) {
	return s.entries, s.err
}

func testScope(t *testing.T) domain.Scope {
	t.Helper()
	scope, err := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")
	if err != nil {
		t.Fatal(err)
	}
	return scope
}

func entry(id, statement string, scope domain.Scope, now time.Time, origin domain.KnowledgeOrigin, status domain.VerificationStatus, evidence []domain.EvidenceReference) domain.LoreEntry {
	if origin == "" {
		origin = domain.KnowledgeOriginHumanAuthored
	}
	if status == "" {
		status = domain.VerificationUnverified
	}
	if evidence == nil {
		evidence = []domain.EvidenceReference{}
	}
	return domain.LoreEntry{
		ID: id, Statement: statement, Scope: scope,
		Origin: origin, VerificationStatus: status,
		Evidence: evidence, CreatedAt: now, UpdatedAt: now,
	}
}

func TestRepositoryProfileRequiresRepositoryScope(t *testing.T) {
	handler := queries.NewRepositoryProfileHandler(&stubLister{}, &stubSearcher{})
	_, err := handler.Handle(context.Background(), queries.RepositoryProfileQuery{})
	if err == nil {
		t.Fatal("expected validation for missing scope")
	}
	team, _ := domain.NewScope(domain.ScopeKindTeam, "payments")
	_, err = handler.Handle(context.Background(), queries.RepositoryProfileQuery{Scope: team})
	if err == nil {
		t.Fatal("expected validation for non-repository kind")
	}
}

func TestRepositoryProfileSectionsAndOmitsEmpty(t *testing.T) {
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	scope := testScope(t)
	list := &stubLister{entries: []domain.LoreEntry{
		entry("d1", "Use Kafka instead of RabbitMQ.", scope, now, domain.KnowledgeOriginHumanAuthored, domain.VerificationVerified, []domain.EvidenceReference{{Type: domain.EvidenceTypeADR, Value: "ADR-017"}}),
		entry("a1", "Hexagonal architecture.", scope, now, "", "", nil),
		entry("u1", "The sky is blue.", scope, now, "", "", nil),
	}}
	handler := queries.NewRepositoryProfileHandler(list, &stubSearcher{})
	result, err := handler.Handle(context.Background(), queries.RepositoryProfileQuery{Scope: scope})
	if err != nil {
		t.Fatal(err)
	}
	if result.Repository.Key != scope.Key {
		t.Fatalf("repository = %+v", result.Repository)
	}
	ids := make([]appcontext.ProfileSectionID, 0, len(result.Sections))
	for _, s := range result.Sections {
		ids = append(ids, s.ID)
		if len(s.Items) == 0 {
			t.Fatalf("empty section %s", s.ID)
		}
	}
	if len(ids) != 2 || ids[0] != appcontext.SectionDecisions || ids[1] != appcontext.SectionArchitecture {
		t.Fatalf("section ids = %v", ids)
	}
	if result.Meta.UnclassifiedCount != 1 {
		t.Fatalf("unclassified = %d", result.Meta.UnclassifiedCount)
	}
	if result.Meta.ItemsIncluded != 2 {
		t.Fatalf("included = %d", result.Meta.ItemsIncluded)
	}
	hasMigrations := false
	for _, s := range result.Sections {
		if s.ID == appcontext.SectionMigrations {
			hasMigrations = true
		}
	}
	if hasMigrations {
		t.Fatal("migrations section should be omitted")
	}
}

func TestRepositoryProfileEmptyRepoSucceeds(t *testing.T) {
	scope := testScope(t)
	handler := queries.NewRepositoryProfileHandler(&stubLister{}, &stubSearcher{})
	result, err := handler.Handle(context.Background(), queries.RepositoryProfileQuery{Scope: scope})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Sections) != 0 {
		t.Fatalf("sections = %+v", result.Sections)
	}
	if result.Meta.ItemsIncluded != 0 {
		t.Fatalf("included = %d", result.Meta.ItemsIncluded)
	}
}

func TestRepositoryProfileOmitsStaleAndSurfacesConflicts(t *testing.T) {
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	scope := testScope(t)
	succ := "succ"
	list := &stubLister{entries: []domain.LoreEntry{
		entry("a", "Use blue-green architecture.", scope, now, "", "", nil),
		entry("b", "Use rolling architecture.", scope, now, "", "", nil),
		{
			ID: "stale", Statement: "Old hexagonal architecture.", Scope: scope,
			Origin: domain.KnowledgeOriginHumanAuthored, VerificationStatus: domain.VerificationVerified,
			SupersededByID: &succ, CreatedAt: now, UpdatedAt: now, Evidence: []domain.EvidenceReference{},
		},
	}}
	handler := queries.NewRepositoryProfileHandler(list, &stubSearcher{})
	result, err := handler.Handle(context.Background(), queries.RepositoryProfileQuery{Scope: scope})
	if err != nil {
		t.Fatal(err)
	}
	for _, sec := range result.Sections {
		for _, item := range sec.Items {
			if item.ID == "stale" {
				t.Fatal("stale item in sections")
			}
		}
	}
	if len(result.Conflicts) == 0 {
		t.Fatal("expected conflicts")
	}
}

func TestRepositoryProfileVerifiedOutranksUnverified(t *testing.T) {
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	scope := testScope(t)
	list := &stubLister{entries: []domain.LoreEntry{
		entry("inf", "Hexagonal architecture inferred.", scope, now, domain.KnowledgeOriginAgentInference, domain.VerificationUnverified, nil),
		entry("hum", "Hexagonal architecture is canonical.", scope, now, domain.KnowledgeOriginHumanAuthored, domain.VerificationVerified, nil),
	}}
	handler := queries.NewRepositoryProfileHandler(list, &stubSearcher{})
	result, err := handler.Handle(context.Background(), queries.RepositoryProfileQuery{Scope: scope})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Sections) != 1 || len(result.Sections[0].Items) < 2 {
		t.Fatalf("sections = %+v", result.Sections)
	}
	if result.Sections[0].Items[0].ID != "hum" {
		t.Fatalf("first item = %s, want hum", result.Sections[0].Items[0].ID)
	}
}
