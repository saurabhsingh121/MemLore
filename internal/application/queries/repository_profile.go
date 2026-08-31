package queries

import (
	"context"
	"strings"
	"time"

	appcontext "github.com/memlore/memlore/internal/application/context"
	"github.com/memlore/memlore/internal/domain"
)

// Default overview query for repository profiles (no caller query in F020).
const DefaultProfileQuery = "architecture decisions conventions technologies ownership gotchas migrations risks dependencies"

// RepositoryProfileQuery is input for a compiled repository briefing.
type RepositoryProfileQuery struct {
	Scope       domain.Scope
	TokenBudget int
}

// RepositoryProfileResult is the on-read repository intelligence profile.
type RepositoryProfileResult struct {
	Repository domain.Scope
	Sections   []appcontext.ProfileSection
	Meta       appcontext.ProfileMeta
	Warnings   []string
	Conflicts  []appcontext.ConflictGroup
}

type loreLister interface {
	Handle(ctx context.Context, query ListLoreByScopeQuery) ([]domain.LoreEntry, error)
}

// RepositoryProfileHandler compiles a token-budgeted repository profile.
type RepositoryProfileHandler struct {
	list   loreLister
	search knowledgeSearcher
	now    func() time.Time
}

// NewRepositoryProfileHandler wires list + search into a profile compiler.
func NewRepositoryProfileHandler(list loreLister, search knowledgeSearcher) *RepositoryProfileHandler {
	return &RepositoryProfileHandler{
		list:   list,
		search: search,
		now:    time.Now,
	}
}

// Handle lists current repository lore, enriches with graph search, ranks,
// classifies, and applies a token budget to named sections only.
func (h *RepositoryProfileHandler) Handle(ctx context.Context, query RepositoryProfileQuery) (RepositoryProfileResult, error) {
	if strings.TrimSpace(query.Scope.Key) == "" {
		return RepositoryProfileResult{}, &domain.ValidationError{Message: "scope is required"}
	}
	if query.Scope.Kind != domain.ScopeKindRepository {
		return RepositoryProfileResult{}, &domain.ValidationError{Message: "scope kind must be repository"}
	}

	budget := query.TokenBudget
	if budget <= 0 {
		budget = appcontext.DefaultTokenBudget
	}

	entries, err := h.list.Handle(ctx, ListLoreByScopeQuery{Scope: query.Scope, IncludeStale: false})
	if err != nil {
		return RepositoryProfileResult{}, err
	}
	current := appcontext.FilterCurrent(entries)
	conflicts := appcontext.DetectConflicts(current)

	searchResult, err := h.search.Handle(ctx, SearchKnowledgeQuery{
		Query:        DefaultProfileQuery,
		Scope:        &query.Scope,
		Limit:        defaultCompileRetrievalLimit,
		IncludeStale: false,
	})
	if err != nil {
		return RepositoryProfileResult{}, err
	}

	now := h.now()
	ranked := appcontext.RankAndDedup(current, searchResult.Graph, query.Scope, now)
	sectionsAll, unclassified := appcontext.Classify(ranked)
	flat := flattenSections(sectionsAll)
	selected, used := appcontext.ApplyTokenBudget(flat, budget)
	sections, _ := appcontext.Classify(selected)

	warnings := searchResult.Warnings
	if warnings == nil {
		warnings = []string{}
	}
	if conflicts == nil {
		conflicts = []appcontext.ConflictGroup{}
	}
	if sections == nil {
		sections = []appcontext.ProfileSection{}
	}

	return RepositoryProfileResult{
		Repository: query.Scope,
		Sections:   sections,
		Meta: appcontext.ProfileMeta{
			TokenBudget:       budget,
			EstimatedTokens:   used,
			ItemsIncluded:     len(selected),
			ItemsTotalRanked:  len(ranked),
			UnclassifiedCount: unclassified,
		},
		Warnings:  warnings,
		Conflicts: conflicts,
	}, nil
}

func flattenSections(sections []appcontext.ProfileSection) []appcontext.RankedItem {
	var out []appcontext.RankedItem
	for _, sec := range sections {
		out = append(out, sec.Items...)
	}
	return out
}
