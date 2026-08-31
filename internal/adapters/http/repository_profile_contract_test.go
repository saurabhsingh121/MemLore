package httpadapter_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRepositoryProfileContract(t *testing.T) {
	server := testClient(t)
	creates := []map[string]any{
		{
			"statement": "Hexagonal architecture.",
			"scope":     map[string]string{"kind": "repository", "key": "github.com/acme/payments"},
		},
		{
			"statement": "Use Kafka instead of RabbitMQ.",
			"scope":     map[string]string{"kind": "repository", "key": "github.com/acme/payments"},
			"evidence":  []map[string]string{{"type": "adr", "value": "ADR-017"}},
		},
	}
	for _, body := range creates {
		raw, _ := json.Marshal(body)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/v1/lore-entries", bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Memlore-Actor", "alice")
		server.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("create status = %d body=%s", rec.Code, rec.Body.String())
		}
	}

	profileBody, _ := json.Marshal(map[string]any{
		"scope": map[string]string{"kind": "repository", "key": "github.com/acme/payments"},
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/repository-profile", bytes.NewReader(profileBody))
	req.Header.Set("Content-Type", "application/json")
	server.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("profile status = %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	repo := resp["repository"].(map[string]any)
	if repo["key"] != "github.com/acme/payments" {
		t.Fatalf("repository = %v", repo)
	}
	sections := resp["sections"].([]any)
	if len(sections) < 2 {
		t.Fatalf("sections = %v", sections)
	}
	ids := map[string]bool{}
	for _, raw := range sections {
		sec := raw.(map[string]any)
		ids[sec["id"].(string)] = true
		items := sec["items"].([]any)
		if len(items) == 0 {
			t.Fatalf("empty section %v", sec["id"])
		}
	}
	if !ids["architecture"] || !ids["decisions"] {
		t.Fatalf("section ids = %v", ids)
	}
	if _, ok := ids["migrations"]; ok {
		t.Fatal("migrations should be omitted")
	}
	meta := resp["meta"].(map[string]any)
	if meta["unclassified_count"] == nil {
		t.Fatal("missing unclassified_count")
	}
	if _, ok := resp["conflicts"].([]any); !ok {
		t.Fatal("conflicts missing")
	}
}

func TestRepositoryProfileRejectsNonRepositoryScope(t *testing.T) {
	server := testClient(t)
	body, _ := json.Marshal(map[string]any{
		"scope": map[string]string{"kind": "team", "key": "payments"},
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/repository-profile", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	server.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestRepositoryProfileEmptyRepoOK(t *testing.T) {
	server := testClient(t)
	body, _ := json.Marshal(map[string]any{
		"scope": map[string]string{"kind": "repository", "key": "github.com/acme/empty"},
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/repository-profile", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	server.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}
