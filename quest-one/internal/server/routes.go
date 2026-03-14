package server

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/quest-one/quest-one/internal/application"
	"github.com/quest-one/quest-one/internal/server/handlers"
	"github.com/quest-one/quest-one/internal/web"
)

func registerRoutes(r chi.Router, app *application.App, log *slog.Logger) {
	h := handlers.New(app, log)

	// Static assets (embedded)
	r.Handle("/static/*", http.StripPrefix("/static", web.StaticHandler()))

	// UI routes (HTMX + templ)
	r.Get("/", h.IndexPage)
	r.Get("/tasks", h.TasksPage)
	r.Get("/tasks/new", h.NewTaskPage)
	r.Get("/tasks/{id}", h.TaskDetailPage)
	r.Get("/settings", h.SettingsPage)
	r.Get("/integrations", h.IntegrationsPage)

	// HTMX partial routes
	r.Post("/tasks", h.CreateTask)
	r.Post("/tasks/{id}/complete", h.CompleteTask)
	r.Post("/tasks/{id}/cancel", h.CancelTask)
	r.Post("/tasks/{id}/promote", h.PromoteTask)
	r.Post("/tasks/{id}/memo", h.AddMemo)
	r.Get("/tasks/search", h.SearchTasks)

	// JSON API routes (for CLI / MCP use)
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/tasks", h.APIListTasks)
		r.Post("/tasks", h.APICreateTask)
		r.Get("/tasks/next", h.APINextTask)
		r.Get("/tasks/candidates", h.APICandidates)
		r.Get("/tasks/{id}", h.APIGetTask)
		r.Post("/tasks/{id}/complete", h.APICompleteTask)
		r.Post("/tasks/{id}/cancel", h.APICancelTask)
		r.Post("/tasks/{id}/promote", h.APIPromoteTask)
		r.Post("/tasks/{id}/memo", h.APIAddMemo)
		r.Get("/tasks/search", h.APISearchTasks)

		r.Get("/integrations", h.APIListIntegrations)
		r.Post("/integrations", h.APICreateIntegration)
		r.Post("/integrations/{id}/enable", h.APIEnableIntegration)
		r.Post("/integrations/{id}/disable", h.APIDisableIntegration)
		r.Delete("/integrations/{id}", h.APIDeleteIntegration)
		r.Post("/sync", h.APISync)

		r.Get("/settings", h.APIGetSettings)
		r.Put("/settings", h.APIUpdateSettings)
	})
}
