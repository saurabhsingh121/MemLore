package queries

import (
	"context"
	"strings"
	"time"

	appcontext "github.com/memlore/memlore/internal/application/context"
	"github.com/memlore/memlore/internal/domain"
)

const defaultCompileRetrievalLimit = appcontext.DefaultRetrievalLimit

// CompileContextQuery is input for context compilation.
type CompileContextQuery struct {
	Task         string
	Query        string
	Scope        domain.Scope
	TokenBudget  int
	Branch       string
	Ticket       string
	ChangedFiles []string
	WorkingFiles []string
	AgentID      string
}

// CompileContextResult is the compiled context packet.
type CompileContextResult struct {
	Task         string
	Query        string
	Scope        domain.Scope
	Branch       string
	Ticket       string
	ChangedFiles []string
	WorkingFiles []string
	AgentID      string
	Items        []appcontext.RankedItem
	Sections     []appcontext.ProfileSection
	Sources      []domain.EvidenceReference
	Meta         appcontext.Meta
	Warnings     []string
	Conflicts    []appcontext.ConflictGroup
}

type knowledgeSearcher interface {
	Handle(ctx context.Context, query SearchKnowledgeQuery) (SearchKnowledgeResult, error)
}

// CompileContextHandler compiles a token-budgeted context packet for agents.
type CompileContextHandler struct {
	search knowledgeSearcher
	list   loreLister
	now    func() time.Time
}

// NewCompileContextHandler wires the compiler with search and optional list-by-scope.
func NewCompileContextHandler(search knowledgeSearcher, list loreLister) *CompileContextHandler {
	return &CompileContextHandler{
		search: search,
		list:   list,
		now:    time.Now,
	}
}

// Handle retrieves knowledge and compiles ranked, budgeted context.
func (h *CompileContextHandler) Handle(ctx context.Context, query CompileContextQuery) (CompileContextResult, error) {
	task := strings.TrimSpace(query.Task)
	if task == "" {
		return CompileContextResult{}, &domain.ValidationError{Message: "task is required"}
	}
	displayQuery := strings.TrimSpace(query.Query)
	if displayQuery == "" {
		displayQuery = task
	}
	if strings.TrimSpace(query.Scope.Key) == "" {
		return CompileContextResult{}, &domain.ValidationError{Message: "scope is required"}
	}

	budget := query.TokenBudget
	if budget <= 0 {
		budget = appcontext.DefaultTokenBudget
	}

	changed := compactStrings(query.ChangedFiles)
	working := compactStrings(query.WorkingFiles)
	ticket := strings.TrimSpace(query.Ticket)
	branch := strings.TrimSpace(query.Branch)
	agentID := strings.TrimSpace(query.AgentID)

	scope := query.Scope
	searchText := compileSearchText(displayQuery, ticket, changed, working)
	searchResult, err := h.search.Handle(ctx, SearchKnowledgeQuery{
		Query:        searchText,
		Scope:        &scope,
		Limit:        defaultCompileRetrievalLimit,
		IncludeStale: false,
	})
	if err != nil {
		return CompileContextResult{}, err
	}

	governance := searchResult.LoreEntries()
	if h.list != nil && scope.Kind == domain.ScopeKindRepository {
		listed, listErr := h.list.Handle(ctx, ListLoreByScopeQuery{Scope: scope, IncludeStale: false})
		if listErr != nil {
			return CompileContextResult{}, listErr
		}
		governance = mergeLoreByID(governance, briefingLore(listed))
	}

	// retrieve → temporal filter → conflict detect → evaluate+rank → budget → classify.
	current := appcontext.FilterCurrent(governance)
	conflicts := appcontext.DetectConflicts(current)

	now := h.now()
	ranked := appcontext.RankAndDedup(current, searchResult.Graph, scope, now)
	selected, used := appcontext.ApplyTokenBudget(ranked, budget)

	sig := appcontext.TaskSignals{
		Task:         task,
		Query:        displayQuery,
		Ticket:       ticket,
		ChangedFiles: changed,
		WorkingFiles: working,
	}
	sections, unclassified := appcontext.ClassifyPacket(selected, sig)
	if sections == nil {
		sections = []appcontext.ProfileSection{}
	}

	warnings := searchResult.Warnings
	if warnings == nil {
		warnings = []string{}
	}
	if conflicts == nil {
		conflicts = []appcontext.ConflictGroup{}
	}

	return CompileContextResult{
		Task:         task,
		Query:        displayQuery,
		Scope:        scope,
		Branch:       branch,
		Ticket:       ticket,
		ChangedFiles: changed,
		WorkingFiles: working,
		AgentID:      agentID,
		Items:        selected,
		Sections:     sections,
		Sources:      uniqueEvidence(selected),
		Meta: appcontext.Meta{
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

func compileSearchText(displayQuery, ticket string, changed, working []string) string {
	parts := make([]string, 0, 2+len(changed)+len(working))
	if displayQuery != "" {
		parts = append(parts, displayQuery)
	}
	if ticket != "" {
		parts = append(parts, ticket)
	}
	parts = append(parts, changed...)
	parts = append(parts, working...)
	return strings.Join(parts, " ")
}

func compactStrings(in []string) []string {
	var out []string
	for _, s := range in {
		if t := strings.TrimSpace(s); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func briefingLore(entries []domain.LoreEntry) []domain.LoreEntry {
	var out []domain.LoreEntry
	for _, e := range entries {
		item := appcontext.RankedItem{
			Statement: e.Statement,
			Evidence:  e.Evidence,
			AuthorityFactors: appcontext.AuthorityFactors{
				Origin: string(e.Origin),
			},
		}
		id, ok := appcontext.ClassifyItem(item)
		if ok && appcontext.IsBriefingSection(id) {
			out = append(out, e)
		}
	}
	return out
}

func mergeLoreByID(primary, extra []domain.LoreEntry) []domain.LoreEntry {
	seen := make(map[string]struct{}, len(primary)+len(extra))
	out := make([]domain.LoreEntry, 0, len(primary)+len(extra))
	for _, e := range primary {
		if e.ID == "" {
			out = append(out, e)
			continue
		}
		if _, ok := seen[e.ID]; ok {
			continue
		}
		seen[e.ID] = struct{}{}
		out = append(out, e)
	}
	for _, e := range extra {
		if e.ID != "" {
			if _, ok := seen[e.ID]; ok {
				continue
			}
			seen[e.ID] = struct{}{}
		}
		out = append(out, e)
	}
	return out
}

func uniqueEvidence(items []appcontext.RankedItem) []domain.EvidenceReference {
	seen := make(map[string]struct{})
	var out []domain.EvidenceReference
	for _, item := range items {
		for _, ev := range item.Evidence {
			key := string(ev.Type) + "\x00" + ev.Value
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, ev)
		}
	}
	return out
}
