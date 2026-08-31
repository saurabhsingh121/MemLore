package context_test

import (
	"testing"

	appcontext "github.com/memlore/memlore/internal/application/context"
	"github.com/memlore/memlore/internal/domain"
)

func TestClassifyItemADREvidenceIsDecision(t *testing.T) {
	item := appcontext.RankedItem{
		Statement: "Use Kafka instead of RabbitMQ.",
		Evidence:  []domain.EvidenceReference{{Type: domain.EvidenceTypeADR, Value: "ADR-017"}},
	}
	id, ok := appcontext.ClassifyItem(item)
	if !ok || id != appcontext.SectionDecisions {
		t.Fatalf("got %q ok=%v, want decisions", id, ok)
	}
}

func TestClassifyItemArchitectureDecisionOrigin(t *testing.T) {
	item := appcontext.RankedItem{
		Statement: "Transactional outbox for all events.",
		AuthorityFactors: appcontext.AuthorityFactors{
			Origin: string(domain.KnowledgeOriginArchitectureDecision),
		},
	}
	id, ok := appcontext.ClassifyItem(item)
	if !ok || id != appcontext.SectionDecisions {
		t.Fatalf("got %q ok=%v, want decisions", id, ok)
	}
}

func TestClassifyItemCues(t *testing.T) {
	cases := []struct {
		statement string
		want      appcontext.ProfileSectionID
		match     bool
	}{
		{"Hexagonal architecture with ports.", appcontext.SectionArchitecture, true},
		{"PostgreSQL is the system of record.", appcontext.SectionTechnologies, true},
		{"Controllers must not access repositories directly.", appcontext.SectionConventions, true},
		{"Gotcha: refunds are eventually consistent.", appcontext.SectionGotchas, true},
		{"Java 17 to Java 21 migration is in progress.", appcontext.SectionMigrations, true},
		{"CODEOWNERS: payments team owns this repo.", appcontext.SectionOwnership, true},
		{"On-call runbook for Kafka outage.", appcontext.SectionOperationalRisks, true},
		{"RefundService is a fragile hotspot.", appcontext.SectionHotspots, true},
		{"payments-api depends on refunds-service.", appcontext.SectionRelatedServices, true},
		{"Recent change: outbox publisher shipped as of 2026-01.", appcontext.SectionRecentChanges, true},
		{"The sky is blue.", "", false},
	}
	for _, tc := range cases {
		id, ok := appcontext.ClassifyItem(appcontext.RankedItem{Statement: tc.statement})
		if ok != tc.match || id != tc.want {
			t.Fatalf("%q: got (%q, %v) want (%q, %v)", tc.statement, id, ok, tc.want, tc.match)
		}
	}
}

func TestClassifyOmitsUnmatchedAndEmptySections(t *testing.T) {
	items := []appcontext.RankedItem{
		{Statement: "The sky is blue."},
		{Statement: "Hexagonal architecture."},
		{Statement: "Use Kafka instead of RabbitMQ.", Evidence: []domain.EvidenceReference{{Type: domain.EvidenceTypeADR, Value: "ADR-017"}}},
	}
	sections, unclassified := appcontext.Classify(items)
	if unclassified != 1 {
		t.Fatalf("unclassified = %d, want 1", unclassified)
	}
	if len(sections) != 2 {
		t.Fatalf("sections = %d, want 2: %+v", len(sections), sections)
	}
	if sections[0].ID != appcontext.SectionDecisions {
		t.Fatalf("first section = %s, want decisions", sections[0].ID)
	}
	if sections[1].ID != appcontext.SectionArchitecture {
		t.Fatalf("second section = %s, want architecture", sections[1].ID)
	}
}

func TestClassifyPacketNamedSectionsAndTaskContext(t *testing.T) {
	items := []appcontext.RankedItem{
		{Statement: "Hexagonal architecture with ports."},
		{Statement: "Use Kafka instead of RabbitMQ.", Evidence: []domain.EvidenceReference{{Type: domain.EvidenceTypeADR, Value: "ADR-017"}}},
		{Statement: "Controllers must not access repositories directly."},
		{Statement: "Gotcha: refunds are eventually consistent."},
		{Statement: "Payment outbox must persist events atomically."},
		{Statement: "The sky is blue."},
	}
	sections, unclassified := appcontext.ClassifyPacket(items, appcontext.TaskSignals{
		Task: "Implement payment outbox handler",
	})
	if unclassified != 1 {
		t.Fatalf("unclassified = %d, want 1", unclassified)
	}
	got := map[appcontext.ProfileSectionID]int{}
	for _, sec := range sections {
		if len(sec.Items) == 0 {
			t.Fatalf("empty section emitted: %s", sec.ID)
		}
		got[sec.ID] = len(sec.Items)
	}
	if got[appcontext.SectionArchitecture] != 1 || got[appcontext.SectionDecisions] != 1 ||
		got[appcontext.SectionConventions] != 1 || got[appcontext.SectionGotchas] != 1 {
		t.Fatalf("briefing sections = %v", got)
	}
	if got[appcontext.SectionTaskContext] != 1 {
		t.Fatalf("task_context = %d, want 1: %v", got[appcontext.SectionTaskContext], sections)
	}
	if _, ok := got[appcontext.SectionTechnologies]; ok {
		t.Fatal("technologies must not be a packet section")
	}
}

func TestClassifyPacketOmitsEmptyAndUsesFiles(t *testing.T) {
	items := []appcontext.RankedItem{
		{Statement: "Publisher writes to src/payments/outbox.go after commit."},
		{Statement: "Unrelated refund SLA is 2 days."},
	}
	sections, unclassified := appcontext.ClassifyPacket(items, appcontext.TaskSignals{
		Task:         "wire publisher",
		ChangedFiles: []string{"src/payments/outbox.go"},
	})
	if unclassified != 1 {
		t.Fatalf("unclassified = %d, want 1", unclassified)
	}
	if len(sections) != 1 || sections[0].ID != appcontext.SectionTaskContext {
		t.Fatalf("sections = %+v", sections)
	}
	if sections[0].Items[0].Statement != items[0].Statement {
		t.Fatalf("task_context item = %+v", sections[0].Items)
	}
}

func TestIsBriefingSection(t *testing.T) {
	if !appcontext.IsBriefingSection(appcontext.SectionArchitecture) ||
		appcontext.IsBriefingSection(appcontext.SectionTaskContext) ||
		appcontext.IsBriefingSection(appcontext.SectionMigrations) {
		t.Fatal("briefing section mapping")
	}
}
