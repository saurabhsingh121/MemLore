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
