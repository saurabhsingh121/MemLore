package httpadapter

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/memlore/memlore/internal/adapters/presenters"
	"github.com/memlore/memlore/internal/application/commands"
	"github.com/memlore/memlore/internal/application/queries"
	"github.com/memlore/memlore/internal/domain"
)

type decisionAlternativeDTO struct {
	Label string `json:"label"`
	Note  string `json:"note,omitempty"`
}

type decisionResponse struct {
	ID                 string                   `json:"id"`
	Scope              presenters.Scope         `json:"scope"`
	Question           string                   `json:"question"`
	Decision           string                   `json:"decision"`
	Rationale          string                   `json:"rationale"`
	Alternatives       []decisionAlternativeDTO `json:"alternatives"`
	Consequences       string                   `json:"consequences"`
	Owner              string                   `json:"owner"`
	DecidedAt          time.Time                `json:"decided_at"`
	AffectedComponents []string                 `json:"affected_components"`
	Evidence           []presenters.Evidence    `json:"evidence"`
	SourceKind         string                   `json:"source_kind"`
	SupersededByID     *string                  `json:"superseded_by_id"`
	Current            bool                     `json:"current"`
	CreatedBy          string                   `json:"created_by"`
	CreatedAt          time.Time                `json:"created_at"`
}

type decisionListResponse struct {
	Items []decisionResponse `json:"items"`
}

type createDecisionRequest struct {
	Scope              presenters.Scope         `json:"scope"`
	Question           string                   `json:"question"`
	Decision           string                   `json:"decision"`
	Rationale          string                   `json:"rationale"`
	Alternatives       []decisionAlternativeDTO `json:"alternatives"`
	Consequences       string                   `json:"consequences"`
	Owner              string                   `json:"owner"`
	DecidedAt          *time.Time               `json:"decided_at"`
	AffectedComponents []string                 `json:"affected_components"`
	Evidence           []presenters.Evidence    `json:"evidence"`
}

func toDecisionResponse(d domain.Decision) decisionResponse {
	alts := make([]decisionAlternativeDTO, 0, len(d.Alternatives))
	for _, a := range d.Alternatives {
		alts = append(alts, decisionAlternativeDTO{Label: a.Label, Note: a.Note})
	}
	components := d.AffectedComponents
	if components == nil {
		components = []string{}
	}
	evidence := make([]presenters.Evidence, 0, len(d.Evidence))
	for _, ev := range d.Evidence {
		evidence = append(evidence, presenters.Evidence{Type: ev.Type, Value: ev.Value})
	}
	return decisionResponse{
		ID: d.ID,
		Scope: presenters.Scope{
			Kind: d.Scope.Kind,
			Key:  d.Scope.Key,
		},
		Question:           d.Question,
		Decision:           d.Choice,
		Rationale:          d.Rationale,
		Alternatives:       alts,
		Consequences:       d.Consequences,
		Owner:              d.Owner,
		DecidedAt:          d.DecidedAt,
		AffectedComponents: components,
		Evidence:           evidence,
		SourceKind:         string(d.SourceKind),
		SupersededByID:     d.SupersededByID,
		Current:            d.Current,
		CreatedBy:          d.CreatedBy,
		CreatedAt:          d.CreatedAt,
	}
}

func parseDecisionEvidence(items []presenters.Evidence) ([]domain.EvidenceReference, error) {
	out := make([]domain.EvidenceReference, 0, len(items))
	for _, item := range items {
		ref, err := domain.NewEvidenceReference(item.Type, item.Value)
		if err != nil {
			return nil, err
		}
		out = append(out, ref)
	}
	return out, nil
}

func parseDecisionAlternatives(items []decisionAlternativeDTO) []domain.DecisionAlternative {
	out := make([]domain.DecisionAlternative, 0, len(items))
	for _, item := range items {
		out = append(out, domain.DecisionAlternative{Label: item.Label, Note: item.Note})
	}
	return out
}

func (h *Handlers) createDecision(w http.ResponseWriter, r *http.Request) {
	actor, err := h.actorFor(r, domain.PermWrite)
	if err != nil {
		handleDomainError(w, err)
		return
	}
	var body createDecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", "invalid JSON body")
		return
	}
	kind, err := domain.ParseScopeKind(string(body.Scope.Kind))
	if err != nil {
		handleDomainError(w, err)
		return
	}
	scope, err := domain.NewScope(kind, body.Scope.Key)
	if err != nil {
		handleDomainError(w, err)
		return
	}
	if err := h.requireScopeAccess(r, scope, false); err != nil {
		handleDomainError(w, err)
		return
	}
	evidence, err := parseDecisionEvidence(body.Evidence)
	if err != nil {
		handleDomainError(w, err)
		return
	}
	cmd := commands.CreateDecisionCommand{
		Scope:              scope,
		Question:           body.Question,
		Choice:             body.Decision,
		Rationale:          body.Rationale,
		Alternatives:       parseDecisionAlternatives(body.Alternatives),
		Consequences:       body.Consequences,
		Owner:              body.Owner,
		AffectedComponents: body.AffectedComponents,
		Evidence:           evidence,
		ActorID:            actor,
	}
	if body.DecidedAt != nil {
		cmd.DecidedAt = *body.DecidedAt
	}
	got, err := h.CreateDecision.Handle(r.Context(), cmd)
	if err != nil {
		handleDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toDecisionResponse(got))
}

func (h *Handlers) getDecision(w http.ResponseWriter, r *http.Request) {
	if err := h.ensureReadAuthorized(r); err != nil {
		handleDomainError(w, err)
		return
	}
	id := chi.URLParam(r, "id")
	got, err := h.GetDecision.Handle(r.Context(), queries.GetDecisionQuery{ID: id})
	if err != nil {
		handleDomainError(w, err)
		return
	}
	if err := h.requireScopeAccess(r, got.Scope, true); err != nil {
		handleDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toDecisionResponse(got))
}

func (h *Handlers) listDecisions(w http.ResponseWriter, r *http.Request) {
	if err := h.ensureReadAuthorized(r); err != nil {
		handleDomainError(w, err)
		return
	}
	scope, err := h.parseIngestScope(r)
	if err != nil {
		handleDomainError(w, err)
		return
	}
	if err := h.requireScopeAccess(r, scope, false); err != nil {
		handleDomainError(w, err)
		return
	}
	items, err := h.ListDecisions.Handle(r.Context(), queries.ListDecisionsQuery{Scope: scope})
	if err != nil {
		handleDomainError(w, err)
		return
	}
	resp := decisionListResponse{Items: make([]decisionResponse, 0, len(items))}
	for _, item := range items {
		resp.Items = append(resp.Items, toDecisionResponse(item))
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handlers) supersedeDecision(w http.ResponseWriter, r *http.Request) {
	actor, err := h.actorFor(r, domain.PermWrite)
	if err != nil {
		handleDomainError(w, err)
		return
	}
	id := chi.URLParam(r, "id")
	existing, err := h.GetDecision.Handle(r.Context(), queries.GetDecisionQuery{ID: id})
	if err != nil {
		handleDomainError(w, err)
		return
	}
	if err := h.requireScopeAccess(r, existing.Scope, true); err != nil {
		handleDomainError(w, err)
		return
	}
	var body createDecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", "invalid JSON body")
		return
	}
	evidence, err := parseDecisionEvidence(body.Evidence)
	if err != nil {
		handleDomainError(w, err)
		return
	}
	cmd := commands.SupersedeDecisionCommand{
		PredecessorID:      id,
		Question:           body.Question,
		Choice:             body.Decision,
		Rationale:          body.Rationale,
		Alternatives:       parseDecisionAlternatives(body.Alternatives),
		Consequences:       body.Consequences,
		Owner:              body.Owner,
		AffectedComponents: body.AffectedComponents,
		Evidence:           evidence,
		ActorID:            actor,
	}
	if body.DecidedAt != nil {
		cmd.DecidedAt = *body.DecidedAt
	}
	got, err := h.SupersedeDecision.Handle(r.Context(), cmd)
	if err != nil {
		handleDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toDecisionResponse(got))
}
