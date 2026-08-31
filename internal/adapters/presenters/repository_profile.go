package presenters

import (
	"github.com/memlore/memlore/internal/application/queries"
)

// ProfileSection is a named non-empty group of profile items.
type ProfileSection struct {
	ID    string        `json:"id"`
	Items []ContextItem `json:"items"`
}

// ProfileMeta summarizes repository profile compilation.
type ProfileMeta struct {
	TokenBudget       int `json:"token_budget"`
	EstimatedTokens   int `json:"estimated_tokens"`
	ItemsIncluded     int `json:"items_included"`
	ItemsTotalRanked  int `json:"items_total_ranked"`
	UnclassifiedCount int `json:"unclassified_count"`
}

// RepositoryProfile is the REST/MCP repository intelligence briefing.
type RepositoryProfile struct {
	Repository Scope            `json:"repository"`
	Sections   []ProfileSection `json:"sections"`
	Meta       ProfileMeta      `json:"meta"`
	Warnings   []string         `json:"warnings"`
	Conflicts  []ConflictGroup  `json:"conflicts"`
}

// ToRepositoryProfile maps handler output to the public profile shape.
func ToRepositoryProfile(result queries.RepositoryProfileResult) RepositoryProfile {
	sections := make([]ProfileSection, 0, len(result.Sections))
	for _, sec := range result.Sections {
		items := make([]ContextItem, 0, len(sec.Items))
		for _, item := range sec.Items {
			items = append(items, toContextItem(item))
		}
		sections = append(sections, ProfileSection{ID: string(sec.ID), Items: items})
	}
	warnings := result.Warnings
	if warnings == nil {
		warnings = []string{}
	}
	return RepositoryProfile{
		Repository: Scope{Kind: result.Repository.Kind, Key: result.Repository.Key},
		Sections:   sections,
		Meta: ProfileMeta{
			TokenBudget:       result.Meta.TokenBudget,
			EstimatedTokens:   result.Meta.EstimatedTokens,
			ItemsIncluded:     result.Meta.ItemsIncluded,
			ItemsTotalRanked:  result.Meta.ItemsTotalRanked,
			UnclassifiedCount: result.Meta.UnclassifiedCount,
		},
		Warnings:  warnings,
		Conflicts: toConflictGroups(result.Conflicts),
	}
}
