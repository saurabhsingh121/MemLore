package context

import (
	"strings"

	"github.com/memlore/memlore/internal/domain"
)

// ProfileSectionID is a named engineering section on a repository profile.
type ProfileSectionID string

const (
	SectionDecisions         ProfileSectionID = "decisions"
	SectionConventions       ProfileSectionID = "conventions"
	SectionGotchas           ProfileSectionID = "gotchas"
	SectionMigrations        ProfileSectionID = "migrations"
	SectionOwnership         ProfileSectionID = "ownership"
	SectionOperationalRisks  ProfileSectionID = "operational_risks"
	SectionHotspots          ProfileSectionID = "hotspots"
	SectionRelatedServices   ProfileSectionID = "related_services"
	SectionArchitecture      ProfileSectionID = "architecture"
	SectionTechnologies      ProfileSectionID = "technologies"
	SectionRecentChanges     ProfileSectionID = "recent_changes"
)

// ClassificationOrder is first-match priority (most specific first).
var ClassificationOrder = []ProfileSectionID{
	SectionDecisions,
	SectionConventions,
	SectionGotchas,
	SectionMigrations,
	SectionOwnership,
	SectionOperationalRisks,
	SectionHotspots,
	SectionRelatedServices,
	SectionArchitecture,
	SectionTechnologies,
	SectionRecentChanges,
}

// ProfileSection is a non-empty group of classified items.
type ProfileSection struct {
	ID    ProfileSectionID
	Items []RankedItem
}

// ProfileMeta summarizes profile compilation.
type ProfileMeta struct {
	TokenBudget       int
	EstimatedTokens   int
	ItemsIncluded     int
	ItemsTotalRanked  int
	UnclassifiedCount int
}

// ClassifyItem returns the first matching section, or false if unmatched.
func ClassifyItem(item RankedItem) (ProfileSectionID, bool) {
	if hasADREvidence(item) || item.AuthorityFactors.Origin == string(domain.KnowledgeOriginArchitectureDecision) {
		return SectionDecisions, true
	}
	text := strings.ToLower(item.Statement)
	switch {
	case containsAny(text, "adr-", "adr ", "decision", "we chose", "instead of"):
		return SectionDecisions, true
	case containsAny(text, "convention", "must not", "must never", "coding standard"):
		return SectionConventions, true
	case containsAny(text, "gotcha", "pitfall", "caveat", "watch out", "never ", "eventually consistent"):
		return SectionGotchas, true
	case strings.Contains(text, "migrat"):
		return SectionMigrations, true
	case containsAny(text, "owner", "owned by", "codeowners"):
		return SectionOwnership, true
	case containsAny(text, "outage", "incident", "on-call", "operational risk"):
		return SectionOperationalRisks, true
	case containsAny(text, "hotspot", "fragile", "frequently changed"):
		return SectionHotspots, true
	case containsAny(text, "depends on", "related service", "publishes", "consumes", "dependency"):
		return SectionRelatedServices, true
	case containsAny(text, "architecture", "hexagonal", "layered", "microservice"):
		return SectionArchitecture, true
	case containsAny(text, "postgres", "postgresql", "kafka", "redis", "neo4j", "java ", "golang", "python "):
		return SectionTechnologies, true
	case containsAny(text, "recent change", "changelog", "as of 20"):
		return SectionRecentChanges, true
	default:
		return "", false
	}
}

// Classify groups items into non-empty sections in ClassificationOrder.
func Classify(items []RankedItem) (sections []ProfileSection, unclassified int) {
	buckets := make(map[ProfileSectionID][]RankedItem, len(ClassificationOrder))
	for _, item := range items {
		id, ok := ClassifyItem(item)
		if !ok {
			unclassified++
			continue
		}
		buckets[id] = append(buckets[id], item)
	}
	for _, id := range ClassificationOrder {
		if got := buckets[id]; len(got) > 0 {
			sections = append(sections, ProfileSection{ID: id, Items: got})
		}
	}
	return sections, unclassified
}

func hasADREvidence(item RankedItem) bool {
	for _, ref := range item.Evidence {
		if ref.Type == domain.EvidenceTypeADR {
			return true
		}
	}
	return false
}

func containsAny(text string, needles ...string) bool {
	for _, n := range needles {
		if strings.Contains(text, n) {
			return true
		}
	}
	return false
}
