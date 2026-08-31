package githubhttp_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/infrastructure/githubhttp"
)

func TestReaderRequiresToken(t *testing.T) {
	r := githubhttp.NewReader("http://example.invalid", "", nil)
	_, err := r.ListPullRequests(context.Background(), ports.PullRequestQuery{Owner: "acme", Repo: "payments"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "token") {
		t.Fatalf("err = %v", err)
	}
}

func TestReaderFetchesMergedPR(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/acme/payments/pulls/1842", func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer secret-token" {
			t.Errorf("authorization = %q", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"number":    1842,
			"node_id":   "PR_kw1",
			"title":     "Use outbox",
			"body":      "because dual-writes race",
			"html_url":  "https://github.com/acme/payments/pull/1842",
			"merged_at": "2026-08-01T12:00:00Z",
			"user":      map[string]string{"login": "dev", "type": "User"},
		})
	})
	mux.HandleFunc("/repos/acme/payments/pulls/1842/files", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]map[string]string{{"filename": "outbox.go"}})
	})
	mux.HandleFunc("/repos/acme/payments/pulls/1842/comments", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]map[string]any{{
			"html_url": "https://github.com/acme/payments/pull/1842#discussion_r1",
			"body":     "because dual-writes race",
			"user":     map[string]string{"login": "reviewer", "type": "User"},
		}})
	})
	mux.HandleFunc("/repos/acme/payments/issues/1842/comments", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]any{})
	})
	mux.HandleFunc("/repos/acme/payments/pulls", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]map[string]any{{
			"number":    1842,
			"merged_at": "2026-08-01T12:00:00Z",
			"title":     "Use outbox",
			"user":      map[string]string{"login": "dev", "type": "User"},
		}})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	r := githubhttp.NewReader(srv.URL, "secret-token", srv.Client())
	prs, err := r.ListPullRequests(context.Background(), ports.PullRequestQuery{Owner: "acme", Repo: "payments", Number: 1842})
	if err != nil {
		t.Fatal(err)
	}
	if len(prs) != 1 || !prs[0].Merged || prs[0].Title != "Use outbox" {
		t.Fatalf("prs = %+v", prs)
	}
	if prs[0].MergedAt == nil || !prs[0].MergedAt.Equal(time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)) {
		t.Fatalf("merged_at = %v", prs[0].MergedAt)
	}
	if len(prs[0].Files) != 1 || prs[0].Files[0] != "outbox.go" {
		t.Fatalf("files = %+v", prs[0].Files)
	}
}

func TestReaderDoesNotLeakTokenInErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"bad creds"}`))
	}))
	t.Cleanup(srv.Close)
	r := githubhttp.NewReader(srv.URL, "super-secret-token-value", srv.Client())
	_, err := r.ListPullRequests(context.Background(), ports.PullRequestQuery{Owner: "acme", Repo: "payments", Number: 1})
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), "super-secret-token-value") {
		t.Fatalf("token leaked: %v", err)
	}
}
