// Package handlers implements HTTP handlers for both UI (HTMX) and JSON API routes.
package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/quest-one/quest-one/internal/application"
)

// Handlers holds all handler dependencies.
type Handlers struct {
	app *application.App
	log *slog.Logger
}

// New constructs a Handlers instance.
func New(app *application.App, log *slog.Logger) *Handlers {
	return &Handlers{app: app, log: log}
}

// ---- helper functions ----

func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func respondError(w http.ResponseWriter, status int, msg string) {
	respondJSON(w, status, map[string]string{"error": msg})
}

func decodeJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

// isHTMX returns true if the request was initiated by HTMX.
func isHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}
