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
	"github.com/memlore/memlore/internal/domain"
	"github.com/memlore/memlore/internal/infrastructure/clock"
	"github.com/memlore/memlore/internal/infrastructure/memory"
)

func decisionServer(t *testing.T) (http.Handler, *memory.UnitOfWork) {
	t.Helper()
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	fixed := clock.FixedClock{Instant: time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)}
	return httpadapter.NewHandlers(begin, fixed, &memory.KnowledgeGraph{}, "test").Router(), uow
}

func TestDecisionRESTContract(t *testing.T) {
	server, uow := decisionServer(t)
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")
	adrEv, _ := domain.NewEvidenceReference(domain.EvidenceTypeADR, "0001-use-postgres")
	adr, _ := domain.NewArchitectureDecisionLoreEntry(domain.NewLoreEntryInput{
		Statement: "Use PostgreSQL as the system of record.",
		Scope:     scope, CreatedBy: "ingest", Evidence: []domain.EvidenceReference{adrEv}, ID: "adr-1",
		Now: time.Date(2026, 9, 1, 11, 0, 0, 0, time.UTC),
	})
	_ = uow.LoreEntries().Add(context.Background(), adr)
	gitEv, _ := domain.NewEvidenceReference(domain.EvidenceTypeCommit, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	git, _ := domain.NewObservationalLoreEntry(domain.NewLoreEntryInput{
		Statement: "Use the outbox because dual-writes race.", Scope: scope, CreatedBy: "ingest",
		Evidence: []domain.EvidenceReference{gitEv},
	})
	_ = uow.LoreEntries().Add(context.Background(), git)

	createBody, _ := json.Marshal(map[string]any{
		"scope":    map[string]string{"kind": "repository", "key": "github.com/acme/payments"},
		"question": "How should payment events be published?",
		"decision": "Transactional outbox",
		"owner":    "alice",
		"alternatives": []map[string]string{
			{"label": "Dual-write to the topic", "note": "Lost updates on crash"},
		},
	})
	createRec := httptest.NewRecorder()
	createReq := httptest.NewRequest(http.MethodPost, "/v1/decisions", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("X-Memlore-Actor", "alice")
	server.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create = %d %s", createRec.Code, createRec.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	id, _ := created["id"].(string)
	if id == "" || created["source_kind"] != "human" || created["decision"] != "Transactional outbox" {
		t.Fatalf("created = %+v", created)
	}

	getRec := httptest.NewRecorder()
	server.ServeHTTP(getRec, httptest.NewRequest(http.MethodGet, "/v1/decisions/"+id, nil))
	if getRec.Code != http.StatusOK {
		t.Fatalf("get = %d %s", getRec.Code, getRec.Body.String())
	}

	adrGet := httptest.NewRecorder()
	server.ServeHTTP(adrGet, httptest.NewRequest(http.MethodGet, "/v1/decisions/adr-1", nil))
	if adrGet.Code != http.StatusOK {
		t.Fatalf("adr get = %d %s", adrGet.Code, adrGet.Body.String())
	}
	var adrBody map[string]any
	_ = json.Unmarshal(adrGet.Body.Bytes(), &adrBody)
	if adrBody["source_kind"] != "adr" {
		t.Fatalf("adr body = %+v", adrBody)
	}

	listRec := httptest.NewRecorder()
	listReq := httptest.NewRequest(http.MethodGet, "/v1/decisions?scope_kind=repository&scope_key=github.com/acme/payments", nil)
	server.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list = %d %s", listRec.Code, listRec.Body.String())
	}
	var listed struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(listRec.Body.Bytes(), &listed); err != nil {
		t.Fatal(err)
	}
	if len(listed.Items) != 2 {
		t.Fatalf("list items = %+v", listed.Items)
	}

	supBody, _ := json.Marshal(map[string]any{
		"question": "How should payment events be published?",
		"decision": "Outbox + idempotent consumers",
		"owner":    "alice",
	})
	supRec := httptest.NewRecorder()
	supReq := httptest.NewRequest(http.MethodPost, "/v1/decisions/"+id+"/supersede", bytes.NewReader(supBody))
	supReq.Header.Set("Content-Type", "application/json")
	supReq.Header.Set("X-Memlore-Actor", "alice")
	server.ServeHTTP(supRec, supReq)
	if supRec.Code != http.StatusCreated {
		t.Fatalf("supersede = %d %s", supRec.Code, supRec.Body.String())
	}

	loreBody, _ := json.Marshal(map[string]any{
		"statement": "A snippet",
		"scope":     map[string]string{"kind": "repository", "key": "github.com/acme/payments"},
	})
	loreRec := httptest.NewRecorder()
	loreReq := httptest.NewRequest(http.MethodPost, "/v1/lore-entries", bytes.NewReader(loreBody))
	loreReq.Header.Set("Content-Type", "application/json")
	loreReq.Header.Set("X-Memlore-Actor", "alice")
	server.ServeHTTP(loreRec, loreReq)
	if loreRec.Code != http.StatusCreated {
		t.Fatalf("remember = %d %s", loreRec.Code, loreRec.Body.String())
	}
	var lore map[string]any
	_ = json.Unmarshal(loreRec.Body.Bytes(), &lore)
	if lore["origin"] != "human_authored" || lore["verification_status"] != "unverified" {
		t.Fatalf("lore-entries must stay unverified snippets: %+v", lore)
	}
}
