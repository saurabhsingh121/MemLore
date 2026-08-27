package httpadapter_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestKnowledgeSearchContract(t *testing.T) {
	server := testClient(t)
	createBody, _ := json.Marshal(map[string]any{
		"statement": "Use outbox for payments.",
		"scope":     map[string]string{"kind": "repository", "key": "github.com/acme/payments"},
	})
	createRec := httptest.NewRecorder()
	createReq := httptest.NewRequest(http.MethodPost, "/v1/lore-entries", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("X-Memlore-Actor", "alice")
	server.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d", createRec.Code)
	}

	searchBody, _ := json.Marshal(map[string]any{
		"query": "payment outbox",
		"scope": map[string]string{"kind": "repository", "key": "github.com/acme/payments"},
		"limit": 10,
	})
	searchRec := httptest.NewRecorder()
	searchReq := httptest.NewRequest(http.MethodPost, "/v1/knowledge-search", bytes.NewReader(searchBody))
	searchReq.Header.Set("Content-Type", "application/json")
	server.ServeHTTP(searchRec, searchReq)
	if searchRec.Code != http.StatusOK {
		t.Fatalf("search status = %d body=%s", searchRec.Code, searchRec.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(searchRec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["query"] != "payment outbox" {
		t.Fatalf("query = %v", resp["query"])
	}
	governance := resp["governance"].(map[string]any)
	if len(governance["items"].([]any)) != 1 {
		t.Fatalf("governance items = %v", governance["items"])
	}
	graph := resp["graph"].(map[string]any)
	if graph["items"] == nil {
		t.Fatal("graph.items should not be null")
	}
	warnings := resp["warnings"].([]any)
	if len(warnings) != 0 {
		t.Fatalf("warnings = %v", warnings)
	}
	for _, forbidden := range []string{"group_id", "EntityEdge", "graphiti"} {
		if bytes.Contains(searchRec.Body.Bytes(), []byte(forbidden)) {
			t.Fatalf("response contains forbidden key %q", forbidden)
		}
	}
}

func TestKnowledgeSearchRejectsEmptyQuery(t *testing.T) {
	server := testClient(t)
	body, _ := json.Marshal(map[string]any{"query": "  "})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/knowledge-search", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	server.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}
