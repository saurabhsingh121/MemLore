package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
)

var _ ports.IngestRepository = (*IngestRepository)(nil)

// IngestRepository is an in-memory ingest store for tests.
type IngestRepository struct {
	mu     sync.RWMutex
	runs   map[string]domain.IngestRun
	cursor map[string]domain.IngestCursor
	shas   map[string]domain.ProcessedSHA
}

func NewIngestRepository() *IngestRepository {
	return &IngestRepository{
		runs:   make(map[string]domain.IngestRun),
		cursor: make(map[string]domain.IngestCursor),
		shas:   make(map[string]domain.ProcessedSHA),
	}
}

func scopeKey(scope domain.Scope) string {
	return string(scope.Kind) + "|" + scope.Key
}

func shaKey(scope domain.Scope, sha string) string {
	return scopeKey(scope) + "|" + sha
}

func (r *IngestRepository) InsertRun(_ context.Context, run domain.IngestRun) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if run.Status == domain.IngestRunRunning {
		sk := scopeKey(run.Scope)
		for _, existing := range r.runs {
			if scopeKey(existing.Scope) == sk && existing.Status == domain.IngestRunRunning {
				return &domain.ConflictError{Message: "ingest already running for repository"}
			}
		}
	}
	r.runs[run.ID] = run
	return nil
}

func (r *IngestRepository) UpdateRun(_ context.Context, run domain.IngestRun) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.runs[run.ID]; !ok {
		return &domain.NotFoundError{Message: "ingest run not found"}
	}
	r.runs[run.ID] = run
	return nil
}

func (r *IngestRepository) GetRun(_ context.Context, id string) (domain.IngestRun, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	run, ok := r.runs[id]
	if !ok {
		return domain.IngestRun{}, &domain.NotFoundError{Message: "ingest run not found"}
	}
	return run, nil
}

func (r *IngestRepository) ListRunsByScope(_ context.Context, scope domain.Scope) ([]domain.IngestRun, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.IngestRun, 0)
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

func (r *IngestRepository) GetCursor(_ context.Context, scope domain.Scope) (domain.IngestCursor, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.cursor[scopeKey(scope)]
	if !ok {
		return domain.IngestCursor{}, false, nil
	}
	return c, true, nil
}

func (r *IngestRepository) UpsertCursor(_ context.Context, cursor domain.IngestCursor) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cursor[scopeKey(cursor.Scope)] = cursor
	return nil
}

func (r *IngestRepository) GetProcessedSHA(_ context.Context, scope domain.Scope, sha string) (domain.ProcessedSHA, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	row, ok := r.shas[shaKey(scope, sha)]
	if !ok {
		return domain.ProcessedSHA{}, false, nil
	}
	return row, true, nil
}

func (r *IngestRepository) InsertProcessedSHA(_ context.Context, row domain.ProcessedSHA) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := shaKey(row.Scope, row.SHA)
	if _, exists := r.shas[key]; exists {
		return fmt.Errorf("processed sha already exists")
	}
	r.shas[key] = row
	return nil
}
