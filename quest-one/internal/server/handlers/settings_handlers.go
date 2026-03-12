package handlers

import (
	"errors"
	"net/http"

	"github.com/quest-one/quest-one/internal/domain"
)

func (h *Handlers) APIGetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.app.GetSettings(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, settings)
}

func (h *Handlers) APIUpdateSettings(w http.ResponseWriter, r *http.Request) {
	var settings domain.Settings
	if err := decodeJSON(r, &settings); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if err := h.app.UpdateSettings(r.Context(), settings); err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, settings)
}
