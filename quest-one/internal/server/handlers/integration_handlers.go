package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/quest-one/quest-one/internal/application"
	"github.com/quest-one/quest-one/internal/domain"
)

func (h *Handlers) APIListIntegrations(w http.ResponseWriter, r *http.Request) {
	integrations, err := h.app.ListIntegrations(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, integrations)
}

func (h *Handlers) APICreateIntegration(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Provider string             `json:"provider"`
		Name     string             `json:"name"`
		BaseURL  string             `json:"base_url"`
		Filters  domain.SyncFilters `json:"filters"`
	}
	if err := decodeJSON(r, &body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	integration, err := h.app.AddIntegration(r.Context(), application.AddIntegrationInput{
		Provider: domain.IntegrationProvider(body.Provider),
		Name:     body.Name,
		BaseURL:  body.BaseURL,
		Filters:  body.Filters,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, integration)
}

func (h *Handlers) APIEnableIntegration(w http.ResponseWriter, r *http.Request) {
	id := domain.IntegrationID(chi.URLParam(r, "id"))
	integration, err := h.app.EnableIntegration(r.Context(), id)
	if err != nil {
		if domain.IsNotFound(err) {
			respondError(w, http.StatusNotFound, "integration not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, integration)
}

func (h *Handlers) APIDisableIntegration(w http.ResponseWriter, r *http.Request) {
	id := domain.IntegrationID(chi.URLParam(r, "id"))
	integration, err := h.app.DisableIntegration(r.Context(), id)
	if err != nil {
		if domain.IsNotFound(err) {
			respondError(w, http.StatusNotFound, "integration not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, integration)
}

func (h *Handlers) APIDeleteIntegration(w http.ResponseWriter, r *http.Request) {
	id := domain.IntegrationID(chi.URLParam(r, "id"))
	if err := h.app.DeleteIntegration(r.Context(), id); err != nil {
		if domain.IsNotFound(err) {
			respondError(w, http.StatusNotFound, "integration not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) APISync(w http.ResponseWriter, r *http.Request) {
	// Source adapters are registered at startup; sync without them returns empty results.
	results, err := h.app.SyncAll(r.Context(), nil)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, results)
}
