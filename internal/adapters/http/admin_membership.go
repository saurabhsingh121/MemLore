package httpadapter

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/memlore/memlore/internal/domain"
)

type adminTeamRequest struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type adminProjectRequest struct {
	Key     string `json:"key"`
	Name    string `json:"name"`
	TeamKey string `json:"team_key"`
}

type adminMemberRequest struct {
	Subject string `json:"subject"`
}

type adminBindingRequest struct {
	ScopeKind  string `json:"scope_kind"`
	ScopeKey   string `json:"scope_key"`
	ProjectKey string `json:"project_key"`
}

func (h *Handlers) requireMembershipStore() error {
	if h.Membership == nil {
		return &domain.ValidationError{Message: "membership store not configured"}
	}
	return nil
}

func (h *Handlers) adminCreateTeam(w http.ResponseWriter, r *http.Request) {
	if err := h.requireAdmin(r); err != nil {
		handleDomainError(w, err)
		return
	}
	if err := h.requireMembershipStore(); err != nil {
		handleDomainError(w, err)
		return
	}
	var body adminTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", "invalid JSON body")
		return
	}
	if err := h.Membership.CreateTeam(r.Context(), body.Key, body.Name); err != nil {
		handleDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"key": body.Key, "name": body.Name})
}

func (h *Handlers) adminCreateProject(w http.ResponseWriter, r *http.Request) {
	if err := h.requireAdmin(r); err != nil {
		handleDomainError(w, err)
		return
	}
	if err := h.requireMembershipStore(); err != nil {
		handleDomainError(w, err)
		return
	}
	var body adminProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", "invalid JSON body")
		return
	}
	if err := h.Membership.CreateProject(r.Context(), body.Key, body.Name, body.TeamKey); err != nil {
		handleDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"key": body.Key, "name": body.Name, "team_key": body.TeamKey})
}

func (h *Handlers) adminAddTeamMember(w http.ResponseWriter, r *http.Request) {
	if err := h.requireAdmin(r); err != nil {
		handleDomainError(w, err)
		return
	}
	if err := h.requireMembershipStore(); err != nil {
		handleDomainError(w, err)
		return
	}
	var body adminMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", "invalid JSON body")
		return
	}
	key := chi.URLParam(r, "key")
	if err := h.Membership.AddTeamMember(r.Context(), key, body.Subject); err != nil {
		handleDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"team_key": key, "subject": body.Subject})
}

func (h *Handlers) adminRemoveTeamMember(w http.ResponseWriter, r *http.Request) {
	if err := h.requireAdmin(r); err != nil {
		handleDomainError(w, err)
		return
	}
	if err := h.requireMembershipStore(); err != nil {
		handleDomainError(w, err)
		return
	}
	key := chi.URLParam(r, "key")
	subject := chi.URLParam(r, "subject")
	if err := h.Membership.RemoveTeamMember(r.Context(), key, subject); err != nil {
		handleDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) adminAddProjectMember(w http.ResponseWriter, r *http.Request) {
	if err := h.requireAdmin(r); err != nil {
		handleDomainError(w, err)
		return
	}
	if err := h.requireMembershipStore(); err != nil {
		handleDomainError(w, err)
		return
	}
	var body adminMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", "invalid JSON body")
		return
	}
	key := chi.URLParam(r, "key")
	if err := h.Membership.AddProjectMember(r.Context(), key, body.Subject); err != nil {
		handleDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"project_key": key, "subject": body.Subject})
}

func (h *Handlers) adminRemoveProjectMember(w http.ResponseWriter, r *http.Request) {
	if err := h.requireAdmin(r); err != nil {
		handleDomainError(w, err)
		return
	}
	if err := h.requireMembershipStore(); err != nil {
		handleDomainError(w, err)
		return
	}
	key := chi.URLParam(r, "key")
	subject := chi.URLParam(r, "subject")
	if err := h.Membership.RemoveProjectMember(r.Context(), key, subject); err != nil {
		handleDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) adminBindScope(w http.ResponseWriter, r *http.Request) {
	if err := h.requireAdmin(r); err != nil {
		handleDomainError(w, err)
		return
	}
	if err := h.requireMembershipStore(); err != nil {
		handleDomainError(w, err)
		return
	}
	var body adminBindingRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", "invalid JSON body")
		return
	}
	if err := h.Membership.BindScope(r.Context(), body.ScopeKind, body.ScopeKey, body.ProjectKey); err != nil {
		handleDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, body)
}

func (h *Handlers) adminUnbindScope(w http.ResponseWriter, r *http.Request) {
	if err := h.requireAdmin(r); err != nil {
		handleDomainError(w, err)
		return
	}
	if err := h.requireMembershipStore(); err != nil {
		handleDomainError(w, err)
		return
	}
	kind := r.URL.Query().Get("scope_kind")
	key := r.URL.Query().Get("scope_key")
	if err := h.Membership.UnbindScope(r.Context(), kind, key); err != nil {
		handleDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
