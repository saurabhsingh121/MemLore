package httpadapter_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCompileContextContract(t *testing.T) {
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

	compileBody, _ := json.Marshal(map[string]any{
		"task":  "Implement payment outbox",
		"query": "payment outbox",
		"scope": map[string]string{"kind": "repository", "key": "github.com/acme/payments"},
	})
	compileRec := httptest.NewRecorder()
	compileReq := httptest.NewRequest(http.MethodPost, "/v1/context/compile", bytes.NewReader(compileBody))
	compileReq.Header.Set("Content-Type", "application/json")
	server.ServeHTTP(compileRec, compileReq)
	if compileRec.Code != http.StatusOK {
		t.Fatalf("compile status = %d body=%s", compileRec.Code, compileRec.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(compileRec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["task"] != "Implement payment outbox" {
		t.Fatalf("task = %v", resp["task"])
	}
	items := resp["items"].([]any)
	if len(items) < 1 {
		t.Fatalf("items = %v", items)
	}
	first := items[0].(map[string]any)
	if first["source"] != "governance" {
		t.Fatalf("source = %v", first["source"])
	}
	if _, ok := first["authority_score"]; !ok {
		t.Fatal("missing authority_score")
	}
	meta := resp["meta"].(map[string]any)
	if meta["token_budget"] == nil {
		t.Fatal("missing token_budget in meta")
	}
}

func TestCompileContextRejectsMissingTask(t *testing.T) {
	server := testClient(t)
	body, _ := json.Marshal(map[string]any{
		"scope": map[string]string{"kind": "repository", "key": "r1"},
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/context/compile", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	server.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}
