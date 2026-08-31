package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
)

var _ ports.PRIngestRepository = (*PRIngestRepository)(nil)

// PRIngestRepository is an in-memory PR ingest store for tests.
type PRIngestRepository struct {
	mu     sync.RWMutex
	runs   map[string]domain.PRIngestRun
	cursor map[string]domain.PRIngestCursor
	prs    map[string]domain.ProcessedPR
}

func NewPRIngestRepository() *PRIngestRepository {
	return &PRIngestRepository{
		runs:   make(map[string]domain.PRIngestRun),
		cursor: make(map[string]domain.PRIngestCursor),
		prs:    make(map[string]domain.ProcessedPR),
	}
}

func prKey(scope domain.Scope, n int) string {
	return scopeKey(scope) + "|" + fmt.Sprintf("%d", n)
}

func (r *PRIngestRepository) InsertRun(_ context.Context, run domain.PRIngestRun) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if run.Status == domain.IngestRunRunning {
		sk := scopeKey(run.Scope)
		for _, existing := range r.runs {
			if scopeKey(existing.Scope) == sk && existing.Status == domain.IngestRunRunning {
				return &domain.ConflictError{Message: "pr ingest already running for repository"}
			}
		}
	}
	r.runs[run.ID] = run
	return nil
}

func (r *PRIngestRepository) UpdateRun(_ context.Context, run domain.PRIngestRun) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.runs[run.ID]; !ok {
		return &domain.NotFoundError{Message: "pr ingest run not found"}
	}
	r.runs[run.ID] = run
	return nil
}

func (r *PRIngestRepository) GetRun(_ context.Context, id string) (domain.PRIngestRun, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	run, ok := r.runs[id]
	if !ok {
		return domain.PRIngestRun{}, &domain.NotFoundError{Message: "pr ingest run not found"}
	}
	return run, nil
}

func (r *PRIngestRepository) ListRunsByScope(_ context.Context, scope domain.Scope) ([]domain.PRIngestRun, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.PRIngestRun, 0)
	for _, run := range r.runs {
		if run.Scope.Kind == scope.Kind && run.Scope.Key == scope.Key {
			out = append(out, run)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].StartedAt.After(out[j].StartedAt)
	})
	return out, nil
}

func (r *PRIngestRepository) GetCursor(_ context.Context, scope domain.Scope) (domain.PRIngestCursor, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.cursor[scopeKey(scope)]
	if !ok {
		return domain.PRIngestCursor{}, false, nil
	}
	return c, true, nil
}

func (r *PRIngestRepository) UpsertCursor(_ context.Context, cursor domain.PRIngestCursor) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cursor[scopeKey(cursor.Scope)] = cursor
	return nil
}

func (r *PRIngestRepository) GetProcessedPR(_ context.Context, scope domain.Scope, prNumber int) (domain.ProcessedPR, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	row, ok := r.prs[prKey(scope, prNumber)]
	if !ok {
		return domain.ProcessedPR{}, false, nil
	}
	return row, true, nil
}

func (r *PRIngestRepository) InsertProcessedPR(_ context.Context, row domain.ProcessedPR) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := prKey(row.Scope, row.PRNumber)
	if _, exists := r.prs[key]; exists {
		return &domain.ConflictError{Message: "pull request already processed"}
	}
	r.prs[key] = row
	return nil
}
