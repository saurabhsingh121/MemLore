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
	conflicts, ok := resp["conflicts"].([]any)
	if !ok {
		t.Fatalf("conflicts missing or wrong type: %v", resp["conflicts"])
	}
	if conflicts == nil {
		t.Fatal("conflicts must be present (empty array ok)")
	}
}

func TestCompileContextSurfacesConflictsAndOmitsStale(t *testing.T) {
	server := testClient(t)
	create := func(statement string) string {
		t.Helper()
		raw, _ := json.Marshal(map[string]any{
			"statement": statement,
			"scope":     map[string]string{"kind": "repository", "key": "r1"},
		})
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/v1/lore-entries", bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Memlore-Actor", "alice")
		server.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("create status = %d body=%s", rec.Code, rec.Body.String())
		}
		var resp map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		return resp["id"].(string)
	}

	idA := create("Use blue-green deploys")
	idB := create("Use rolling deploys")
	oldID := create("Legacy deploy rule")
	supBody, _ := json.Marshal(map[string]any{"statement": "Successor deploy rule"})
	supRec := httptest.NewRecorder()
	supReq := httptest.NewRequest(http.MethodPost, "/v1/lore-entries/"+oldID+"/supersede", bytes.NewReader(supBody))
	supReq.Header.Set("Content-Type", "application/json")
	supReq.Header.Set("X-Memlore-Actor", "alice")
	server.ServeHTTP(supRec, supReq)
	if supRec.Code != http.StatusCreated {
		t.Fatalf("supersede status = %d body=%s", supRec.Code, supRec.Body.String())
	}

	compileBody, _ := json.Marshal(map[string]any{
		"task":  "Choose deploy strategy",
		"scope": map[string]string{"kind": "repository", "key": "r1"},
	})
	compileRec := httptest.NewRecorder()
	compileReq := httptest.NewRequest(http.MethodPost, "/v1/context/compile", bytes.NewReader(compileBody))
	compileReq.Header.Set("Content-Type", "application/json")
	server.ServeHTTP(compileRec, compileReq)
	if compileRec.Code != http.StatusOK {
		t.Fatalf("compile status = %d body=%s", compileRec.Code, compileRec.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(compileRec.Body.Bytes(), &resp)
	itemIDs := map[string]bool{}
	for _, raw := range resp["items"].([]any) {
		itemIDs[raw.(map[string]any)["id"].(string)] = true
	}
	if itemIDs[oldID] {
		t.Fatal("superseded predecessor must not appear in items")
	}
	if !itemIDs[idA] || !itemIDs[idB] {
		t.Fatalf("expected both current sides in items: %v", itemIDs)
	}
	conflicts := resp["conflicts"].([]any)
	if len(conflicts) < 1 {
		t.Fatalf("expected conflicts, got %v", conflicts)
	}

	listedRec := httptest.NewRecorder()
	listedReq := httptest.NewRequest(http.MethodGet, "/v1/lore-entries?scope_kind=repository&scope_key=r1", nil)
	server.ServeHTTP(listedRec, listedReq)
	var listed map[string]any
	_ = json.Unmarshal(listedRec.Body.Bytes(), &listed)
	for _, raw := range listed["items"].([]any) {
		if raw.(map[string]any)["id"].(string) == oldID {
			t.Fatal("default list must omit superseded")
		}
	}

	staleRec := httptest.NewRecorder()
	staleReq := httptest.NewRequest(http.MethodGet, "/v1/lore-entries?scope_kind=repository&scope_key=r1&include_stale=true", nil)
	server.ServeHTTP(staleRec, staleReq)
	var staleListed map[string]any
	_ = json.Unmarshal(staleRec.Body.Bytes(), &staleListed)
	foundStale := false
	for _, raw := range staleListed["items"].([]any) {
		if raw.(map[string]any)["id"].(string) == oldID {
			foundStale = true
		}
	}
	if !foundStale {
		t.Fatal("include_stale=true must return superseded")
	}

	getRec := httptest.NewRecorder()
	getReq := httptest.NewRequest(http.MethodGet, "/v1/lore-entries/"+oldID, nil)
	server.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get stale status = %d", getRec.Code)
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
