package memory

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
)

var _ ports.ADRIngestRepository = (*ADRIngestRepository)(nil)

// ADRIngestRepository is an in-memory ADR ingest store for tests.
type ADRIngestRepository struct {
	mu     sync.RWMutex
	runs   map[string]domain.ADRIngestRun
	cursor map[string]domain.ADRIngestCursor
	files  map[string]domain.ProcessedADR
}

func NewADRIngestRepository() *ADRIngestRepository {
	return &ADRIngestRepository{
		runs:   make(map[string]domain.ADRIngestRun),
		cursor: make(map[string]domain.ADRIngestCursor),
		files:  make(map[string]domain.ProcessedADR),
	}
}

func adrFileKey(scope domain.Scope, path, checksum string) string {
	return scopeKey(scope) + "|" + path + "|" + checksum
}

func (r *ADRIngestRepository) InsertRun(_ context.Context, run domain.ADRIngestRun) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if run.Status == domain.IngestRunRunning {
		sk := scopeKey(run.Scope)
		for _, existing := range r.runs {
			if scopeKey(existing.Scope) == sk && existing.Status == domain.IngestRunRunning {
				return &domain.ConflictError{Message: "adr ingest already running for repository"}
			}
		}
	}
	r.runs[run.ID] = run
	return nil
}

func (r *ADRIngestRepository) UpdateRun(_ context.Context, run domain.ADRIngestRun) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.runs[run.ID]; !ok {
		return &domain.NotFoundError{Message: "adr ingest run not found"}
	}
	r.runs[run.ID] = run
	return nil
}

func (r *ADRIngestRepository) GetRun(_ context.Context, id string) (domain.ADRIngestRun, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	run, ok := r.runs[id]
	if !ok {
		return domain.ADRIngestRun{}, &domain.NotFoundError{Message: "adr ingest run not found"}
	}
	return run, nil
}

func (r *ADRIngestRepository) ListRunsByScope(_ context.Context, scope domain.Scope) ([]domain.ADRIngestRun, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.ADRIngestRun, 0)
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

func (r *ADRIngestRepository) GetCursor(_ context.Context, scope domain.Scope) (domain.ADRIngestCursor, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.cursor[scopeKey(scope)]
	if !ok {
		return domain.ADRIngestCursor{}, false, nil
	}
	return c, true, nil
}

func (r *ADRIngestRepository) UpsertCursor(_ context.Context, cursor domain.ADRIngestCursor) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cursor[scopeKey(cursor.Scope)] = cursor
	return nil
}

func (r *ADRIngestRepository) GetProcessedADR(_ context.Context, scope domain.Scope, relativePath, checksum string) (domain.ProcessedADR, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	row, ok := r.files[adrFileKey(scope, relativePath, checksum)]
	if !ok {
		return domain.ProcessedADR{}, false, nil
	}
	return row, true, nil
}

func (r *ADRIngestRepository) InsertProcessedADR(_ context.Context, row domain.ProcessedADR) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := adrFileKey(row.Scope, row.RelativePath, row.Checksum)
	if _, exists := r.files[key]; exists {
		return &domain.ConflictError{Message: "adr file already processed"}
	}
	r.files[key] = row
	return nil
}

func (r *ADRIngestRepository) LatestStoredByPath(_ context.Context, scope domain.Scope, relativePath string) (domain.ProcessedADR, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return latestStored(r.files, func(row domain.ProcessedADR) bool {
		return row.Scope.Kind == scope.Kind && row.Scope.Key == scope.Key && row.RelativePath == relativePath
	})
}

func (r *ADRIngestRepository) LatestStoredByADRID(_ context.Context, scope domain.Scope, adrID string) (domain.ProcessedADR, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return latestStored(r.files, func(row domain.ProcessedADR) bool {
		if row.Scope.Kind != scope.Kind || row.Scope.Key != scope.Key {
			return false
		}
		if row.ADRID == adrID {
			return true
		}
		return strings.HasPrefix(row.ADRID, adrID+"-")
	})
}

func latestStored(files map[string]domain.ProcessedADR, match func(domain.ProcessedADR) bool) (domain.ProcessedADR, bool, error) {
	var best domain.ProcessedADR
	found := false
	var bestAt time.Time
	for _, row := range files {
		if !match(row) || row.Skipped || row.LoreEntryID == "" {
			continue
		}
		if !found || row.ProcessedAt.After(bestAt) {
			best = row
			bestAt = row.ProcessedAt
			found = true
		}
	}
	if !found {
		return domain.ProcessedADR{}, false, nil
	}
	return best, true, nil
}
