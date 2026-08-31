package httpadapter

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/memlore/memlore/internal/adapters/presenters"
	"github.com/memlore/memlore/internal/application/commands"
	"github.com/memlore/memlore/internal/application/queries"
	"github.com/memlore/memlore/internal/domain"
)

type reviewQueueItemResponse struct {
	presenters.LoreEntry
	SourceType  string  `json:"source_type"`
	Status      string  `json:"status,omitempty"`
	SuccessorID *string `json:"successor_id,omitempty"`
}

type reviewQueueListResponse struct {
	Items []reviewQueueItemResponse `json:"items"`
}

type acceptReviewRequest struct {
	Statement *string `json:"statement"`
}

type rejectReviewResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Actor  string `json:"actor_id"`
}

func toReviewQueueItem(item queries.SuggestedLoreItem, includeStatus bool) reviewQueueItemResponse {
	resp := reviewQueueItemResponse{
		LoreEntry:  presenters.ToLoreEntry(item.Entry),
		SourceType: item.SourceType,
	}
	if includeStatus {
		resp.Status = item.Status
		resp.SuccessorID = item.SuccessorID
	}
	return resp
}

func (h *Handlers) listReviewQueue(w http.ResponseWriter, r *http.Request) {
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
	items, err := h.ListReviewQueue.Handle(r.Context(), queries.ListReviewQueueQuery{Scope: scope})
	if err != nil {
		handleDomainError(w, err)
		return
	}
	resp := reviewQueueListResponse{Items: make([]reviewQueueItemResponse, 0, len(items))}
	for _, item := range items {
		resp.Items = append(resp.Items, toReviewQueueItem(item, false))
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handlers) getReviewItem(w http.ResponseWriter, r *http.Request) {
	if err := h.ensureReadAuthorized(r); err != nil {
		handleDomainError(w, err)
		return
	}
	id := chi.URLParam(r, "id")
	item, err := h.GetReviewItem.Handle(r.Context(), queries.GetReviewItemQuery{ID: id})
	if err != nil {
		handleDomainError(w, err)
		return
	}
	if err := h.requireScopeAccess(r, item.Entry.Scope, true); err != nil {
		handleDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toReviewQueueItem(item, true))
}

func (h *Handlers) acceptReviewItem(w http.ResponseWriter, r *http.Request) {
	actor, err := h.actorFor(r, domain.PermWrite)
	if err != nil {
		handleDomainError(w, err)
		return
	}
	id := chi.URLParam(r, "id")
	var body acceptReviewRequest
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil && !errors.Is(err, io.EOF) {
			handleDomainError(w, &domain.ValidationError{Message: "invalid json"})
			return
		}
	}
	item, err := h.GetReviewItem.Handle(r.Context(), queries.GetReviewItemQuery{ID: id})
	if err != nil {
		handleDomainError(w, err)
		return
	}
	if err := h.requireScopeAccess(r, item.Entry.Scope, true); err != nil {
		handleDomainError(w, err)
		return
	}
	succ, err := h.AcceptReview.Handle(r.Context(), commands.AcceptReviewCommand{
		EntryID:   id,
		ActorID:   actor,
		Statement: body.Statement,
	})
	if err != nil {
		handleDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, presenters.ToLoreEntry(succ))
}

func (h *Handlers) rejectReviewItem(w http.ResponseWriter, r *http.Request) {
	actor, err := h.actorFor(r, domain.PermWrite)
	if err != nil {
		handleDomainError(w, err)
		return
	}
	id := chi.URLParam(r, "id")
	item, err := h.GetReviewItem.Handle(r.Context(), queries.GetReviewItemQuery{ID: id})
	if err != nil {
		handleDomainError(w, err)
		return
	}
	if err := h.requireScopeAccess(r, item.Entry.Scope, true); err != nil {
		handleDomainError(w, err)
		return
	}
	got, err := h.RejectReview.Handle(r.Context(), commands.RejectReviewCommand{EntryID: id, ActorID: actor})
	if err != nil {
		handleDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rejectReviewResponse{ID: got.ID, Status: string(got.Status), Actor: got.Actor})
}
