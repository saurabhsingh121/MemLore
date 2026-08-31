package httpadapter_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	httpadapter "github.com/memlore/memlore/internal/adapters/http"
	"github.com/memlore/memlore/internal/domain"
	"github.com/memlore/memlore/internal/infrastructure/clock"
	"github.com/memlore/memlore/internal/infrastructure/memory"
)

// Characterization: tests/contract/test_create_lore_entry.py, test_lore_flow.py

func testClient(t *testing.T) http.Handler {
	t.Helper()
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	fixed := clock.FixedClock{Instant: time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)}
	graph := &memory.KnowledgeGraph{}
	return httpadapter.NewHandlers(begin, fixed, graph, "test").Router()
}

func TestCreateLoreEntryContract(t *testing.T) {
	server := testClient(t)
	body := map[string]any{
		"statement": "Payment events must use the transactional outbox.",
		"scope":     map[string]string{"kind": "repository", "key": "github.com/acme/payments"},
		"evidence":  []map[string]string{{"type": "adr", "value": "0001-dual-plane"}},
	}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/v1/lore-entries", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Memlore-Actor", "alice")
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["origin"] != "human_authored" {
		t.Fatalf("origin = %v", resp["origin"])
	}
	if resp["verification_status"] != "unverified" {
		t.Fatalf("status = %v", resp["verification_status"])
	}
	if resp["created_by"] != "alice" {
		t.Fatalf("created_by = %v", resp["created_by"])
	}
	scope := resp["scope"].(map[string]any)
	if scope["kind"] != "repository" {
		t.Fatalf("scope kind = %v", scope["kind"])
	}
}

func TestCreateRejectsMissingActorAndOversizedStatement(t *testing.T) {
	server := testClient(t)

	missing := httptest.NewRecorder()
	reqMissing := httptest.NewRequest(http.MethodPost, "/v1/lore-entries", bytes.NewReader([]byte(`{"statement":"x","scope":{"kind":"team","key":"t1"}}`)))
	reqMissing.Header.Set("Content-Type", "application/json")
	server.ServeHTTP(missing, reqMissing)
	if missing.Code != http.StatusBadRequest {
		t.Fatalf("missing actor status = %d", missing.Code)
	}

	oversizedBody, _ := json.Marshal(map[string]any{
		"statement": strings.Repeat("x", domain.MaxStatementLength+1),
		"scope":     map[string]string{"kind": "team", "key": "t1"},
	})
	oversized := httptest.NewRecorder()
	reqOversized := httptest.NewRequest(http.MethodPost, "/v1/lore-entries", bytes.NewReader(oversizedBody))
	reqOversized.Header.Set("Content-Type", "application/json")
	reqOversized.Header.Set("X-Memlore-Actor", "alice")
	server.ServeHTTP(oversized, reqOversized)
	if oversized.Code != http.StatusBadRequest {
		t.Fatalf("oversized status = %d", oversized.Code)
	}
}

func TestCreateDuplicateStatementAllowed(t *testing.T) {
	server := testClient(t)
	payload := map[string]any{
		"statement": "Same",
		"scope":     map[string]string{"kind": "team", "key": "t1"},
	}
	raw, _ := json.Marshal(payload)

	post := func() string {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/v1/lore-entries", bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Memlore-Actor", "alice")
		server.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("create status = %d", rec.Code)
		}
		var resp map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		return resp["id"].(string)
	}

	idA := post()
	idB := post()
	if idA == idB {
		t.Fatal("expected distinct ids for duplicate statements")
	}
}

func TestGetVerifyAuditsFlow(t *testing.T) {
	server := testClient(t)
	createBody, _ := json.Marshal(map[string]any{
		"statement": "Rule",
		"scope":     map[string]string{"kind": "repository", "key": "r1"},
	})
	createRec := httptest.NewRecorder()
	createReq := httptest.NewRequest(http.MethodPost, "/v1/lore-entries", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("X-Memlore-Actor", "alice")
	server.ServeHTTP(createRec, createReq)
	var created map[string]any
	_ = json.Unmarshal(createRec.Body.Bytes(), &created)
	entryID := created["id"].(string)

	gotRec := httptest.NewRecorder()
	gotReq := httptest.NewRequest(http.MethodGet, "/v1/lore-entries/"+entryID, nil)
	server.ServeHTTP(gotRec, gotReq)
	if gotRec.Code != http.StatusOK {
		t.Fatalf("get status = %d", gotRec.Code)
	}

	missingRec := httptest.NewRecorder()
	missingReq := httptest.NewRequest(http.MethodGet, "/v1/lore-entries/00000000-0000-0000-0000-000000000000", nil)
	server.ServeHTTP(missingRec, missingReq)
	if missingRec.Code != http.StatusNotFound {
		t.Fatalf("missing status = %d", missingRec.Code)
	}

	verify := func(actor string) map[string]any {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/v1/lore-entries/"+entryID+"/verify", nil)
		req.Header.Set("X-Memlore-Actor", actor)
		server.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("verify status = %d body=%s", rec.Code, rec.Body.String())
		}
		var resp map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		return resp
	}
	first := verify("alice")
	if first["verification_status"] != "verified" || first["origin"] != "human_authored" {
		t.Fatalf("first verify = %v", first)
	}
	second := verify("bob")
	if second["verified_by"] != "alice" {
		t.Fatalf("verified_by = %v", second["verified_by"])
	}

	auditsRec := httptest.NewRecorder()
	auditsReq := httptest.NewRequest(http.MethodGet, "/v1/lore-entries/"+entryID+"/audits", nil)
	server.ServeHTTP(auditsRec, auditsReq)
	var audits map[string]any
	_ = json.Unmarshal(auditsRec.Body.Bytes(), &audits)
	items := audits["items"].([]any)
	createCount, verifyCount := 0, 0
	for _, item := range items {
		action := item.(map[string]any)["action"].(string)
		if action == "create" {
			createCount++
		}
		if action == "verify" {
			verifyCount++
		}
	}
	if createCount != 1 || verifyCount != 1 {
		t.Fatalf("audits = %v", items)
	}

	auditsMissingRec := httptest.NewRecorder()
	auditsMissingReq := httptest.NewRequest(http.MethodGet, "/v1/lore-entries/00000000-0000-0000-0000-000000000000/audits", nil)
	server.ServeHTTP(auditsMissingRec, auditsMissingReq)
	if auditsMissingRec.Code != http.StatusNotFound {
		t.Fatalf("audits missing status = %d", auditsMissingRec.Code)
	}
}

func TestListByScope(t *testing.T) {
	server := testClient(t)
	for _, payload := range []map[string]any{
		{"statement": "Rule", "scope": map[string]string{"kind": "repository", "key": "r1"}},
		{"statement": "Other", "scope": map[string]string{"kind": "team", "key": "r1"}},
	} {
		raw, _ := json.Marshal(payload)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/v1/lore-entries", bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Memlore-Actor", "alice")
		server.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("create status = %d", rec.Code)
		}
	}

	listedRec := httptest.NewRecorder()
	listedReq := httptest.NewRequest(http.MethodGet, "/v1/lore-entries?scope_kind=repository&scope_key=r1", nil)
	server.ServeHTTP(listedRec, listedReq)
	var listed map[string]any
	_ = json.Unmarshal(listedRec.Body.Bytes(), &listed)
	if len(listed["items"].([]any)) != 1 {
		t.Fatalf("listed = %v", listed)
	}

	emptyRec := httptest.NewRecorder()
	emptyReq := httptest.NewRequest(http.MethodGet, "/v1/lore-entries?scope_kind=repository&scope_key=none", nil)
	server.ServeHTTP(emptyRec, emptyReq)
	var empty map[string]any
	_ = json.Unmarshal(emptyRec.Body.Bytes(), &empty)
	if len(empty["items"].([]any)) != 0 {
		t.Fatalf("empty = %v", empty)
	}
}

func TestInvalidateAndSupersedeHTTPContract(t *testing.T) {
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

	entryID := create("Use the outbox")
	invRec := httptest.NewRecorder()
	invReq := httptest.NewRequest(http.MethodPost, "/v1/lore-entries/"+entryID+"/invalidate", nil)
	invReq.Header.Set("X-Memlore-Actor", "alice")
	server.ServeHTTP(invRec, invReq)
	if invRec.Code != http.StatusOK {
		t.Fatalf("invalidate status = %d body=%s", invRec.Code, invRec.Body.String())
	}
	var inv map[string]any
	_ = json.Unmarshal(invRec.Body.Bytes(), &inv)
	if inv["verification_status"] != "invalidated" {
		t.Fatalf("status = %v", inv["verification_status"])
	}

	again := httptest.NewRecorder()
	againReq := httptest.NewRequest(http.MethodPost, "/v1/lore-entries/"+entryID+"/invalidate", nil)
	againReq.Header.Set("X-Memlore-Actor", "bob")
	server.ServeHTTP(again, againReq)
	if again.Code != http.StatusOK {
		t.Fatalf("re-invalidate status = %d", again.Code)
	}

	missingActor := httptest.NewRecorder()
	server.ServeHTTP(missingActor, httptest.NewRequest(http.MethodPost, "/v1/lore-entries/"+entryID+"/invalidate", nil))
	if missingActor.Code != http.StatusBadRequest {
		t.Fatalf("missing actor status = %d", missingActor.Code)
	}

	missing := httptest.NewRecorder()
	missingReq := httptest.NewRequest(http.MethodPost, "/v1/lore-entries/00000000-0000-0000-0000-000000000000/invalidate", nil)
	missingReq.Header.Set("X-Memlore-Actor", "alice")
	server.ServeHTTP(missing, missingReq)
	if missing.Code != http.StatusNotFound {
		t.Fatalf("unknown invalidate status = %d", missing.Code)
	}

	predID := create("Old rule")
	body, _ := json.Marshal(map[string]any{
		"statement": "New rule",
		"evidence":  []map[string]string{{"type": "adr", "value": "0003-lifecycle"}},
	})
	supRec := httptest.NewRecorder()
	supReq := httptest.NewRequest(http.MethodPost, "/v1/lore-entries/"+predID+"/supersede", bytes.NewReader(body))
	supReq.Header.Set("Content-Type", "application/json")
	supReq.Header.Set("X-Memlore-Actor", "bob")
	server.ServeHTTP(supRec, supReq)
	if supRec.Code != http.StatusCreated {
		t.Fatalf("supersede status = %d body=%s", supRec.Code, supRec.Body.String())
	}
	var successor map[string]any
	_ = json.Unmarshal(supRec.Body.Bytes(), &successor)
	if successor["statement"] != "New rule" {
		t.Fatalf("successor = %v", successor)
	}

	gotRec := httptest.NewRecorder()
	server.ServeHTTP(gotRec, httptest.NewRequest(http.MethodGet, "/v1/lore-entries/"+predID, nil))
	var pred map[string]any
	_ = json.Unmarshal(gotRec.Body.Bytes(), &pred)
	if pred["superseded_by_id"] != successor["id"] {
		t.Fatalf("superseded_by_id = %v", pred["superseded_by_id"])
	}

	retry := httptest.NewRecorder()
	retryReq := httptest.NewRequest(http.MethodPost, "/v1/lore-entries/"+predID+"/supersede", bytes.NewReader(body))
	retryReq.Header.Set("Content-Type", "application/json")
	retryReq.Header.Set("X-Memlore-Actor", "bob")
	server.ServeHTTP(retry, retryReq)
	if retry.Code != http.StatusBadRequest {
		t.Fatalf("re-supersede status = %d", retry.Code)
	}

	invSup := httptest.NewRecorder()
	invSupReq := httptest.NewRequest(http.MethodPost, "/v1/lore-entries/"+entryID+"/supersede", bytes.NewReader(body))
	invSupReq.Header.Set("Content-Type", "application/json")
	invSupReq.Header.Set("X-Memlore-Actor", "alice")
	server.ServeHTTP(invSup, invSupReq)
	if invSup.Code != http.StatusBadRequest {
		t.Fatalf("supersede invalidated status = %d", invSup.Code)
	}

	auditsRec := httptest.NewRecorder()
	server.ServeHTTP(auditsRec, httptest.NewRequest(http.MethodGet, "/v1/lore-entries/"+predID+"/audits", nil))
	var audits map[string]any
	_ = json.Unmarshal(auditsRec.Body.Bytes(), &audits)
	items := audits["items"].([]any)
	if items[len(items)-1].(map[string]any)["action"] != "supersede" {
		t.Fatalf("audits = %v", items)
	}
}
