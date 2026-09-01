package context

import (
	"path"
	"strings"
	"unicode"

	"github.com/memlore/memlore/internal/domain"
)

// ProfileSectionID is a named engineering section on a repository profile.
type ProfileSectionID string

const (
	SectionDecisions        ProfileSectionID = "decisions"
	SectionConventions      ProfileSectionID = "conventions"
	SectionGotchas          ProfileSectionID = "gotchas"
	SectionMigrations       ProfileSectionID = "migrations"
	SectionOwnership        ProfileSectionID = "ownership"
	SectionOperationalRisks ProfileSectionID = "operational_risks"
	SectionHotspots         ProfileSectionID = "hotspots"
	SectionRelatedServices  ProfileSectionID = "related_services"
	SectionArchitecture     ProfileSectionID = "architecture"
	SectionTechnologies     ProfileSectionID = "technologies"
	SectionRecentChanges    ProfileSectionID = "recent_changes"
	SectionTaskContext      ProfileSectionID = "task_context"
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
	if item.FirstClassDecision || hasADREvidence(item) || item.AuthorityFactors.Origin == string(domain.KnowledgeOriginArchitectureDecision) {
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

// PacketSectionOrder is display order for get_for_task / compile packet sections.
var PacketSectionOrder = []ProfileSectionID{
	SectionArchitecture,
	SectionDecisions,
	SectionConventions,
	SectionTaskContext,
	SectionGotchas,
}

// TaskSignals are optional compile inputs used to decide task_context membership.
type TaskSignals struct {
	Task         string
	Query        string
	Ticket       string
	ChangedFiles []string
	WorkingFiles []string
}

// IsBriefingSection reports whether id is one of the four repository briefing types
// merged into a task packet (architecture, decisions, conventions, gotchas).
func IsBriefingSection(id ProfileSectionID) bool {
	switch id {
	case SectionArchitecture, SectionDecisions, SectionConventions, SectionGotchas:
		return true
	default:
		return false
	}
}

// ItemTaskRelevant reports whether an item matches task/query/ticket/file needles.
func ItemTaskRelevant(item RankedItem, sig TaskSignals) bool {
	needles := relevanceNeedles(sig)
	if len(needles) == 0 {
		return false
	}
	text := strings.ToLower(item.Statement)
	for _, ref := range item.Evidence {
		text += " " + strings.ToLower(ref.Value)
	}
	return containsAny(text, needles...)
}

// ClassifyPacket groups budgeted items into packet sections. Briefing classes
// keep their F020 ids. Leftover task-relevant items go to task_context.
// Empty sections are omitted. Unmatched non-relevant items increment unclassified.
func ClassifyPacket(items []RankedItem, sig TaskSignals) (sections []ProfileSection, unclassified int) {
	buckets := make(map[ProfileSectionID][]RankedItem, len(PacketSectionOrder))
	for _, item := range items {
		id, ok := ClassifyItem(item)
		if ok && IsBriefingSection(id) {
			buckets[id] = append(buckets[id], item)
			continue
		}
		if ItemTaskRelevant(item, sig) {
			buckets[SectionTaskContext] = append(buckets[SectionTaskContext], item)
			continue
		}
		unclassified++
	}
	for _, id := range PacketSectionOrder {
		if got := buckets[id]; len(got) > 0 {
			sections = append(sections, ProfileSection{ID: id, Items: got})
		}
	}
	return sections, unclassified
}

func relevanceNeedles(sig TaskSignals) []string {
	seen := make(map[string]struct{})
	var needles []string
	add := func(s string) {
		s = strings.ToLower(strings.TrimSpace(s))
		if s == "" {
			return
		}
		if _, ok := seen[s]; ok {
			return
		}
		seen[s] = struct{}{}
		needles = append(needles, s)
	}
	addTokens := func(s string) {
		for _, w := range strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsNumber(r)
		}) {
			if len(w) >= 4 {
				add(w)
			}
		}
	}
	addTokens(sig.Task)
	addTokens(sig.Query)
	addTokens(sig.Ticket)
	for _, f := range append(append([]string{}, sig.ChangedFiles...), sig.WorkingFiles...) {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		add(strings.ToLower(f))
		if base := path.Base(f); base != "" {
			add(strings.ToLower(base))
		}
	}
	return needles
}
