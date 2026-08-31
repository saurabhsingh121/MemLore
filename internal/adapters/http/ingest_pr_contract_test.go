package httpadapter_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	httpadapter "github.com/memlore/memlore/internal/adapters/http"
	"github.com/memlore/memlore/internal/application/commands"
	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/application/queries"
	"github.com/memlore/memlore/internal/domain"
	"github.com/memlore/memlore/internal/infrastructure/clock"
	"github.com/memlore/memlore/internal/infrastructure/memory"
)

type stubPR struct {
	prs []domain.PullRequestSnapshot
}

func (s *stubPR) ListPullRequests(_ context.Context, q ports.PullRequestQuery) ([]domain.PullRequestSnapshot, error) {
	if q.Number > 0 {
		for _, pr := range s.prs {
			if pr.Number == q.Number {
				return []domain.PullRequestSnapshot{pr}, nil
			}
		}
		return nil, nil
	}
	return s.prs, nil
}

func prIngestServer(t *testing.T, prs ports.PullRequestReader) http.Handler {
	t.Helper()
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	fixed := clock.FixedClock{Instant: time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)}
	h := httpadapter.NewHandlers(begin, fixed, &memory.KnowledgeGraph{}, "test")
	h.IngestPR = commands.NewIngestPullRequestsHandler(begin, fixed, prs)
	h.ListPRIngestRuns = queries.NewListPRIngestRunsHandler(begin)
	h.GetPRIngestRun = queries.NewGetPRIngestRunHandler(begin)
	return h.Router()
}

func TestIngestPRContract(t *testing.T) {
	merged := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	prs := &stubPR{prs: []domain.PullRequestSnapshot{{
		Number: 1842, Owner: "acme", Repo: "payments",
		Title: "Use outbox", Body: "because dual-writes race",
		Merged: true, MergedAt: &merged, AuthorLogin: "dev",
	}}}
	server := prIngestServer(t, prs)
	body, _ := json.Marshal(map[string]any{
		"scope": map[string]string{"kind": "repository", "key": "github.com/acme/payments"},
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/ingest/pr", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Memlore-Actor", "alice")
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var run map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &run); err != nil {
		t.Fatal(err)
	}
	if run["status"] != "succeeded" || run["candidates_stored"].(float64) != 1 {
		t.Fatalf("run = %+v", run)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/v1/ingest/pr-runs?scope_kind=repository&scope_key=github.com/acme/payments", nil)
	listRec := httptest.NewRecorder()
	server.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status = %d %s", listRec.Code, listRec.Body.String())
	}

	getReq := httptest.NewRequest(http.MethodGet, "/v1/ingest/pr-runs/"+run["id"].(string), nil)
	getRec := httptest.NewRecorder()
	server.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get status = %d %s", getRec.Code, getRec.Body.String())
	}

	candReq := httptest.NewRequest(http.MethodGet, "/v1/ingest/candidates?scope_kind=repository&scope_key=github.com/acme/payments&evidence_type=pr", nil)
	candRec := httptest.NewRecorder()
	server.ServeHTTP(candRec, candReq)
	if candRec.Code != http.StatusOK {
		t.Fatalf("candidates status = %d %s", candRec.Code, candRec.Body.String())
	}
	var cands map[string]any
	_ = json.Unmarshal(candRec.Body.Bytes(), &cands)
	items := cands["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("candidates = %+v", cands)
	}
	item := items[0].(map[string]any)
	if item["origin"] != "repository_observation" || item["verification_status"] != "unverified" {
		t.Fatalf("item = %+v", item)
	}
}

func TestIngestPRRequiresActorAndValidScope(t *testing.T) {
	server := prIngestServer(t, &stubPR{})
	body, _ := json.Marshal(map[string]any{
		"scope": map[string]string{"kind": "repository", "key": "github.com/acme/payments"},
	})
	missing := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/ingest/pr", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	server.ServeHTTP(missing, req)
	if missing.Code != http.StatusBadRequest {
		t.Fatalf("missing actor = %d", missing.Code)
	}

	bad, _ := json.Marshal(map[string]any{
		"scope": map[string]string{"kind": "team", "key": "t1"},
	})
	teamRec := httptest.NewRecorder()
	teamReq := httptest.NewRequest(http.MethodPost, "/v1/ingest/pr", bytes.NewReader(bad))
	teamReq.Header.Set("Content-Type", "application/json")
	teamReq.Header.Set("X-Memlore-Actor", "alice")
	server.ServeHTTP(teamRec, teamReq)
	if teamRec.Code != http.StatusBadRequest {
		t.Fatalf("team scope = %d %s", teamRec.Code, teamRec.Body.String())
	}
}

func TestIngestGitContractStillWorksAlongsidePR(t *testing.T) {
	git := &stubGit{commits: []domain.GitCommitSnapshot{{
		SHA:         "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		CommittedAt: time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC),
		Subject:     "feat: outbox",
		Body:        "because dual-writes race",
		Message:     "feat: outbox\n\nbecause dual-writes race",
	}}}
	server := ingestServer(t, git)
	body, _ := json.Marshal(map[string]any{
		"scope": map[string]string{"kind": "repository", "key": "github.com/acme/payments"},
		"path":  "/tmp/payments",
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/ingest/git", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Memlore-Actor", "alice")
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("git status = %d %s", rec.Code, rec.Body.String())
	}
}
