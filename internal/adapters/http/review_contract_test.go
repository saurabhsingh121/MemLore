package httpadapter_test

import (
	"bytes"
	"context"
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

func reviewServer(t *testing.T) (http.Handler, *memory.UnitOfWork) {
	t.Helper()
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	fixed := clock.FixedClock{Instant: time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)}
	h := httpadapter.NewHandlers(begin, fixed, &memory.KnowledgeGraph{}, "test")
	return h.Router(), uow
}

func seedReviewObservation(t *testing.T, uow *memory.UnitOfWork, statement string, ev domain.EvidenceReference) domain.LoreEntry {
	t.Helper()
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")
	entry, err := domain.NewObservationalLoreEntry(domain.NewLoreEntryInput{
		Statement: statement,
		Scope:     scope,
		CreatedBy: "ingest",
		Evidence:  []domain.EvidenceReference{ev},
		Now:       time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := uow.LoreEntries().Add(context.Background(), entry); err != nil {
		t.Fatal(err)
	}
	return entry
}

func TestReviewQueueContract(t *testing.T) {
	server, uow := reviewServer(t)
	gitEv, _ := domain.NewEvidenceReference(domain.EvidenceTypeCommit, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	prEv, _ := domain.NewEvidenceReference(domain.EvidenceTypePR, "acme/payments#1842")
	adrEv, _ := domain.NewEvidenceReference(domain.EvidenceTypeADR, "0001-use-postgres")
	git := seedReviewObservation(t, uow, "Use the outbox because dual-writes race.", gitEv)
	pr := seedReviewObservation(t, uow, "Payment events use transactional outbox.", prEv)
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")
	adr, _ := domain.NewArchitectureDecisionLoreEntry(domain.NewLoreEntryInput{
		Statement: "Use PostgreSQL as the system of record.",
		Scope:     scope,
		CreatedBy: "alice",
		Evidence:  []domain.EvidenceReference{adrEv},
	})
	_ = uow.LoreEntries().Add(context.Background(), adr)

	listReq := httptest.NewRequest(http.MethodGet, "/v1/review-queue?scope_kind=repository&scope_key=github.com/acme/payments", nil)
	listRec := httptest.NewRecorder()
	server.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list = %d %s", listRec.Code, listRec.Body.String())
	}
	raw := listRec.Body.String()
	if strings.Contains(raw, `"confidence"`) || strings.Contains(raw, `"reason"`) {
		t.Fatalf("invented fields: %s", raw)
	}
	var listed struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(listRec.Body.Bytes(), &listed); err != nil {
		t.Fatal(err)
	}
	if len(listed.Items) != 2 {
		t.Fatalf("items = %+v", listed.Items)
	}
	for _, item := range listed.Items {
		if item["origin"] != "repository_observation" {
			t.Fatalf("item = %+v", item)
		}
		if item["id"] == adr.ID {
			t.Fatal("ADR in review queue")
		}
	}

	getReq := httptest.NewRequest(http.MethodGet, "/v1/review-queue/"+pr.ID, nil)
	getRec := httptest.NewRecorder()
	server.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get = %d %s", getRec.Code, getRec.Body.String())
	}

	acceptBody, _ := json.Marshal(map[string]any{})
	accReq := httptest.NewRequest(http.MethodPost, "/v1/review-queue/"+git.ID+"/accept", bytes.NewReader(acceptBody))
	accReq.Header.Set("Content-Type", "application/json")
	accReq.Header.Set("X-Memlore-Actor", "alice")
	accRec := httptest.NewRecorder()
	server.ServeHTTP(accRec, accReq)
	if accRec.Code != http.StatusOK {
		t.Fatalf("accept = %d %s", accRec.Code, accRec.Body.String())
	}
	var accepted map[string]any
	if err := json.Unmarshal(accRec.Body.Bytes(), &accepted); err != nil {
		t.Fatal(err)
	}
	if accepted["origin"] != "human_verified" || accepted["verification_status"] != "verified" {
		t.Fatalf("accepted = %+v", accepted)
	}

	editBody, _ := json.Marshal(map[string]any{"statement": "Payment events MUST use the transactional outbox."})
	editReq := httptest.NewRequest(http.MethodPost, "/v1/review-queue/"+pr.ID+"/accept", bytes.NewReader(editBody))
	editReq.Header.Set("Content-Type", "application/json")
	editReq.Header.Set("X-Memlore-Actor", "alice")
	editRec := httptest.NewRecorder()
	server.ServeHTTP(editRec, editReq)
	if editRec.Code != http.StatusOK {
		t.Fatalf("edit accept = %d %s", editRec.Code, editRec.Body.String())
	}
	var edited map[string]any
	_ = json.Unmarshal(editRec.Body.Bytes(), &edited)
	if edited["origin"] != "human_authored" {
		t.Fatalf("edited = %+v", edited)
	}

	extraEv, _ := domain.NewEvidenceReference(domain.EvidenceTypeCommit, "cccccccccccccccccccccccccccccccccccccccc")
	extra := seedReviewObservation(t, uow, "Skip this extract", extraEv)
	rejReq := httptest.NewRequest(http.MethodPost, "/v1/review-queue/"+extra.ID+"/reject", nil)
	rejReq.Header.Set("X-Memlore-Actor", "alice")
	rejRec := httptest.NewRecorder()
	server.ServeHTTP(rejRec, rejReq)
	if rejRec.Code != http.StatusOK {
		t.Fatalf("reject = %d %s", rejRec.Code, rejRec.Body.String())
	}

	list2 := httptest.NewRecorder()
	server.ServeHTTP(list2, httptest.NewRequest(http.MethodGet, "/v1/review-queue?scope_kind=repository&scope_key=github.com/acme/payments", nil))
	var listed2 struct {
		Items []map[string]any `json:"items"`
	}
	_ = json.Unmarshal(list2.Body.Bytes(), &listed2)
	if len(listed2.Items) != 0 {
		t.Fatalf("pending after accept/reject = %+v", listed2.Items)
	}

	candReq := httptest.NewRequest(http.MethodGet, "/v1/ingest/candidates?scope_kind=repository&scope_key=github.com/acme/payments", nil)
	candRec := httptest.NewRecorder()
	server.ServeHTTP(candRec, candReq)
	if candRec.Code != http.StatusOK {
		t.Fatalf("candidates = %d %s", candRec.Code, candRec.Body.String())
	}
}
