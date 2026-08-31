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
	if first["trust_band"] == nil || first["trust_band"] == "" {
		t.Fatal("missing trust_band")
	}
	factors, ok := first["authority_factors"].(map[string]any)
	if !ok {
		t.Fatalf("authority_factors = %v", first["authority_factors"])
	}
	if factors["source_type"] == nil {
		t.Fatal("missing source_type in authority_factors")
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

func TestCompileContextContractSectionsAndV1Body(t *testing.T) {
	server := testClient(t)
	create := func(statement string, evidence []map[string]string) {
		t.Helper()
		body := map[string]any{
			"statement": statement,
			"scope":     map[string]string{"kind": "repository", "key": "github.com/acme/payments"},
		}
		if evidence != nil {
			body["evidence"] = evidence
		}
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
	create("Hexagonal architecture with ports.", nil)
	create("Use Kafka instead of RabbitMQ.", []map[string]string{{"type": "adr", "value": "ADR-017"}})
	create("Payment outbox must persist events atomically.", nil)

	v1Body, _ := json.Marshal(map[string]any{
		"task":  "Implement payment outbox",
		"scope": map[string]string{"kind": "repository", "key": "github.com/acme/payments"},
	})
	v1Rec := httptest.NewRecorder()
	v1Req := httptest.NewRequest(http.MethodPost, "/v1/context/compile", bytes.NewReader(v1Body))
	v1Req.Header.Set("Content-Type", "application/json")
	server.ServeHTTP(v1Rec, v1Req)
	if v1Rec.Code != http.StatusOK {
		t.Fatalf("v1 compile status = %d body=%s", v1Rec.Code, v1Rec.Body.String())
	}
	var v1 map[string]any
	if err := json.Unmarshal(v1Rec.Body.Bytes(), &v1); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"task", "query", "scope", "items", "meta", "warnings", "conflicts"} {
		if _, ok := v1[key]; !ok {
			t.Fatalf("v1 packet missing %s", key)
		}
	}

	richBody, _ := json.Marshal(map[string]any{
		"task":          "Implement payment outbox",
		"scope":         map[string]string{"kind": "repository", "key": "github.com/acme/payments"},
		"changed_files": []string{"src/payments/outbox.go"},
		"ticket":        "PAY-1842",
		"agent_id":      "cursor-agent",
	})
	richRec := httptest.NewRecorder()
	richReq := httptest.NewRequest(http.MethodPost, "/v1/context/compile", bytes.NewReader(richBody))
	richReq.Header.Set("Content-Type", "application/json")
	server.ServeHTTP(richRec, richReq)
	if richRec.Code != http.StatusOK {
		t.Fatalf("rich compile status = %d body=%s", richRec.Code, richRec.Body.String())
	}
	var rich map[string]any
	if err := json.Unmarshal(richRec.Body.Bytes(), &rich); err != nil {
		t.Fatal(err)
	}
	if rich["agent_id"] != "cursor-agent" || rich["ticket"] != "PAY-1842" {
		t.Fatalf("echo fields = %v", rich)
	}
	sections, ok := rich["sections"].([]any)
	if !ok || len(sections) < 2 {
		t.Fatalf("sections = %v", rich["sections"])
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
	if _, ok := ids["conventions"]; ok {
		t.Fatal("empty conventions must be omitted")
	}
	sources, ok := rich["sources"].([]any)
	if !ok || len(sources) < 1 {
		t.Fatalf("sources = %v", rich["sources"])
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
