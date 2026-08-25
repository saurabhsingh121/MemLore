package httpadapter

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/memlore/memlore/internal/adapters/presenters"
	"github.com/memlore/memlore/internal/domain"
)

type errorBody struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Details []any  `json:"details"`
	} `json:"error"`
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	body := errorBody{}
	body.Error.Code = code
	body.Error.Message = message
	body.Error.Details = []any{}
	_ = json.NewEncoder(w).Encode(body)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func requireActor(r *http.Request) (string, error) {
	actor := strings.TrimSpace(r.Header.Get("X-Memlore-Actor"))
	if actor == "" {
		return "", &domain.ValidationError{Message: "X-Memlore-Actor header is required"}
	}
	return actor, nil
}

func handleDomainError(w http.ResponseWriter, err error) {
	switch e := err.(type) {
	case *domain.ValidationError:
		writeError(w, http.StatusBadRequest, "validation_error", e.Message)
	case *domain.NotFoundError:
		writeError(w, http.StatusNotFound, "not_found", e.Message)
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "unexpected error")
	}
}

type scopeDTO = presenters.Scope

type evidenceDTO = presenters.Evidence

type createLoreRequest struct {
	Statement string        `json:"statement"`
	Scope     scopeDTO      `json:"scope"`
	Evidence  []evidenceDTO `json:"evidence"`
}

type loreEntryResponse = presenters.LoreEntry

type loreEntryListResponse = presenters.LoreEntryList

type auditRecordResponse = presenters.AuditRecord

type auditListResponse = presenters.AuditList

func toLoreResponse(entry domain.LoreEntry) loreEntryResponse {
	return presenters.ToLoreEntry(entry)
}

func toAuditResponse(record domain.AuditRecord) auditRecordResponse {
	return presenters.ToAuditRecord(record)
}

func parseCreateRequest(body createLoreRequest) (string, domain.Scope, []domain.EvidenceReference, error) {
	kind, err := domain.ParseScopeKind(string(body.Scope.Kind))
	if err != nil {
		return "", domain.Scope{}, nil, err
	}
	scope, err := domain.NewScope(kind, body.Scope.Key)
	if err != nil {
		return "", domain.Scope{}, nil, err
	}
	evidence := make([]domain.EvidenceReference, 0, len(body.Evidence))
	for _, item := range body.Evidence {
		ref, refErr := domain.NewEvidenceReference(item.Type, item.Value)
		if refErr != nil {
			return "", domain.Scope{}, nil, refErr
		}
		evidence = append(evidence, ref)
	}
	return body.Statement, scope, evidence, nil
}
