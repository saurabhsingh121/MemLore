package authority

import (
	"time"

	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
)

// EvaluateGovernance evaluates a lore entry against the requested compile/explain scope.
func EvaluateGovernance(entry domain.LoreEntry, requested *domain.Scope, now time.Time) domain.Evaluation {
	return domain.EvaluateAuthority(domain.FactorInputs{
		Origin:             entry.Origin,
		VerificationStatus: entry.VerificationStatus,
		Superseded:         domain.IsSuperseded(entry),
		CreatedAt:          entry.CreatedAt,
		Now:                now,
		Evidence:           entry.Evidence,
		EntryScope:         entry.Scope,
		RequestedScope:     requested,
	})
}

// EvaluateGraph evaluates a graph-plane fact.
func EvaluateGraph(fact ports.GraphFact, requested *domain.Scope, now time.Time) domain.Evaluation {
	in := domain.FactorInputs{
		FromGraph:      true,
		GraphScore:     fact.Score,
		Now:            now,
		RequestedScope: requested,
	}
	if fact.Scope != nil {
		kind, err := domain.ParseScopeKind(fact.Scope.Kind)
		if err == nil {
			scope, err := domain.NewScope(kind, fact.Scope.Key)
			if err == nil {
				in.EntryScope = scope
			}
		}
	}
	return domain.EvaluateAuthority(in)
}
