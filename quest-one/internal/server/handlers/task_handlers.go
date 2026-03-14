package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/quest-one/quest-one/internal/application"
	"github.com/quest-one/quest-one/internal/domain"
	"github.com/quest-one/quest-one/internal/web"
)

// ---- UI handlers ----

func (h *Handlers) IndexPage(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/tasks", http.StatusFound)
}

func (h *Handlers) TasksPage(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.app.Candidates(r.Context(), 50)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	_ = web.TasksPage(tasks).Render(r.Context(), w)
}

func (h *Handlers) NewTaskPage(w http.ResponseWriter, r *http.Request) {
	_ = web.NewTaskPage().Render(r.Context(), w)
}

func (h *Handlers) TaskDetailPage(w http.ResponseWriter, r *http.Request) {
	id := domain.TaskID(chi.URLParam(r, "id"))
	task, err := h.app.Tasks.FindByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	_ = web.TaskDetailPage(task).Render(r.Context(), w)
}

func (h *Handlers) SettingsPage(w http.ResponseWriter, r *http.Request) {
	settings, err := h.app.GetSettings(r.Context())
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	_ = web.SettingsPage(settings).Render(r.Context(), w)
}

func (h *Handlers) IntegrationsPage(w http.ResponseWriter, r *http.Request) {
	integrations, err := h.app.ListIntegrations(r.Context())
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	_ = web.IntegrationsPage(integrations).Render(r.Context(), w)
}

// ---- HTMX partial handlers ----

func (h *Handlers) CreateTask(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	in := application.AddTaskInput{
		Title:       r.FormValue("title"),
		Description: r.FormValue("description"),
	}

	task, err := h.app.AddTask(r.Context(), in)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if isHTMX(r) {
		_ = web.TaskRow(task).Render(r.Context(), w)
		return
	}
	http.Redirect(w, r, "/tasks", http.StatusSeeOther)
}

func (h *Handlers) CompleteTask(w http.ResponseWriter, r *http.Request) {
	id := domain.TaskID(chi.URLParam(r, "id"))
	task, err := h.app.CompleteTask(r.Context(), id)
	if err != nil {
		handleTaskError(w, err)
		return
	}
	if isHTMX(r) {
		_ = web.TaskRow(task).Render(r.Context(), w)
		return
	}
	http.Redirect(w, r, "/tasks", http.StatusSeeOther)
}

func (h *Handlers) CancelTask(w http.ResponseWriter, r *http.Request) {
	id := domain.TaskID(chi.URLParam(r, "id"))
	task, err := h.app.CancelTask(r.Context(), id)
	if err != nil {
		handleTaskError(w, err)
		return
	}
	if isHTMX(r) {
		_ = web.TaskRow(task).Render(r.Context(), w)
		return
	}
	http.Redirect(w, r, "/tasks", http.StatusSeeOther)
}

func (h *Handlers) PromoteTask(w http.ResponseWriter, r *http.Request) {
	id := domain.TaskID(chi.URLParam(r, "id"))
	task, err := h.app.PromoteTask(r.Context(), id)
	if err != nil {
		handleTaskError(w, err)
		return
	}
	if isHTMX(r) {
		_ = web.TaskRow(task).Render(r.Context(), w)
		return
	}
	http.Redirect(w, r, "/tasks", http.StatusSeeOther)
}

func (h *Handlers) AddMemo(w http.ResponseWriter, r *http.Request) {
	id := domain.TaskID(chi.URLParam(r, "id"))
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	task, err := h.app.AddMemo(r.Context(), id, r.FormValue("memo"))
	if err != nil {
		handleTaskError(w, err)
		return
	}
	if isHTMX(r) {
		_ = web.TaskDetailPage(task).Render(r.Context(), w)
		return
	}
	http.Redirect(w, r, "/tasks/"+string(id), http.StatusSeeOther)
}

func (h *Handlers) SearchTasks(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	results, err := h.app.SearchTasks(r.Context(), q, 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	tasks := make([]domain.Task, len(results))
	for i, res := range results {
		tasks[i] = res.Task
	}
	if isHTMX(r) {
		_ = web.TaskList(tasks).Render(r.Context(), w)
		return
	}
	_ = web.TasksPage(tasks).Render(r.Context(), w)
}

// ---- JSON API handlers ----

func (h *Handlers) APIListTasks(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit == 0 {
		limit = 50
	}

	out, err := h.app.ListTasks(r.Context(), application.ListTasksInput{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, out)
}

func (h *Handlers) APICreateTask(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
		DueDate     *string  `json:"due_date"`
	}
	if err := decodeJSON(r, &body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	in := application.AddTaskInput{
		Title:       body.Title,
		Description: body.Description,
		Tags:        body.Tags,
	}
	if body.DueDate != nil {
		t, err := time.Parse(time.RFC3339, *body.DueDate)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid due_date format (RFC3339 required)")
			return
		}
		in.DueDate = &t
	}

	task, err := h.app.AddTask(r.Context(), in)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, task)
}

func (h *Handlers) APINextTask(w http.ResponseWriter, r *http.Request) {
	task, err := h.app.NextTask(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if task == nil {
		respondJSON(w, http.StatusOK, nil)
		return
	}
	respondJSON(w, http.StatusOK, task)
}

func (h *Handlers) APICandidates(w http.ResponseWriter, r *http.Request) {
	n, _ := strconv.Atoi(r.URL.Query().Get("n"))
	if n == 0 {
		n = 5
	}
	tasks, err := h.app.Candidates(r.Context(), n)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, tasks)
}

func (h *Handlers) APIGetTask(w http.ResponseWriter, r *http.Request) {
	id := domain.TaskID(chi.URLParam(r, "id"))
	task, err := h.app.Tasks.FindByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			respondError(w, http.StatusNotFound, "task not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, task)
}

func (h *Handlers) APICompleteTask(w http.ResponseWriter, r *http.Request) {
	id := domain.TaskID(chi.URLParam(r, "id"))
	task, err := h.app.CompleteTask(r.Context(), id)
	if err != nil {
		handleTaskErrorJSON(w, err)
		return
	}
	respondJSON(w, http.StatusOK, task)
}

func (h *Handlers) APICancelTask(w http.ResponseWriter, r *http.Request) {
	id := domain.TaskID(chi.URLParam(r, "id"))
	task, err := h.app.CancelTask(r.Context(), id)
	if err != nil {
		handleTaskErrorJSON(w, err)
		return
	}
	respondJSON(w, http.StatusOK, task)
}

func (h *Handlers) APIPromoteTask(w http.ResponseWriter, r *http.Request) {
	id := domain.TaskID(chi.URLParam(r, "id"))
	task, err := h.app.PromoteTask(r.Context(), id)
	if err != nil {
		handleTaskErrorJSON(w, err)
		return
	}
	respondJSON(w, http.StatusOK, task)
}

func (h *Handlers) APIAddMemo(w http.ResponseWriter, r *http.Request) {
	id := domain.TaskID(chi.URLParam(r, "id"))
	var body struct {
		Memo string `json:"memo"`
	}
	if err := decodeJSON(r, &body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	task, err := h.app.AddMemo(r.Context(), id, body.Memo)
	if err != nil {
		handleTaskErrorJSON(w, err)
		return
	}
	respondJSON(w, http.StatusOK, task)
}

func (h *Handlers) APISearchTasks(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	results, err := h.app.SearchTasks(r.Context(), q, limit)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, results)
}

// ---- error helpers ----

func handleTaskError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		http.NotFound(w, nil)
	case errors.Is(err, domain.ErrConflict):
		http.Error(w, err.Error(), http.StatusConflict)
	default:
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func handleTaskErrorJSON(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		respondError(w, http.StatusNotFound, "task not found")
	case errors.Is(err, domain.ErrConflict):
		respondError(w, http.StatusConflict, err.Error())
	default:
		respondError(w, http.StatusInternalServerError, err.Error())
	}
}
