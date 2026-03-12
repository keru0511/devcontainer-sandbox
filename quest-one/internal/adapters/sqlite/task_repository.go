package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/quest-one/quest-one/internal/domain"
	"github.com/quest-one/quest-one/internal/ports"
)

// TaskRepository implements ports.TaskRepository against SQLite.
type TaskRepository struct {
	db  *sql.DB
	log *slog.Logger
}

// Ensure interface compliance at compile time.
var _ ports.TaskRepository = (*TaskRepository)(nil)

func (r *TaskRepository) Save(ctx context.Context, t domain.Task) error {
	tagsJSON, err := json.Marshal(t.Tags)
	if err != nil {
		return fmt.Errorf("task_repository: marshal tags: %w", err)
	}
	priorityJSON, err := json.Marshal(t.Priority)
	if err != nil {
		return fmt.Errorf("task_repository: marshal priority: %w", err)
	}

	var dueDate *string
	if t.DueDate != nil {
		s := t.DueDate.UTC().Format(time.RFC3339)
		dueDate = &s
	}
	var completedAt *string
	if t.CompletedAt != nil {
		s := t.CompletedAt.UTC().Format(time.RFC3339)
		completedAt = &s
	}

	var parentID *string
	if t.ParentID != nil {
		s := string(*t.ParentID)
		parentID = &s
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO tasks
			(id, title, description, status, memo, project_id, parent_id, due_date,
			 tags, priority_json, created_at, updated_at, completed_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(id) DO UPDATE SET
			title        = excluded.title,
			description  = excluded.description,
			status       = excluded.status,
			memo         = excluded.memo,
			project_id   = excluded.project_id,
			parent_id    = excluded.parent_id,
			due_date     = excluded.due_date,
			tags         = excluded.tags,
			priority_json = excluded.priority_json,
			updated_at   = excluded.updated_at,
			completed_at = excluded.completed_at
	`,
		string(t.ID), t.Title, t.Description, string(t.Status), t.Memo,
		t.ProjectID, parentID, dueDate,
		string(tagsJSON), string(priorityJSON),
		t.CreatedAt.UTC().Format(time.RFC3339),
		t.UpdatedAt.UTC().Format(time.RFC3339),
		completedAt,
	)
	return err
}

func (r *TaskRepository) FindByID(ctx context.Context, id domain.TaskID) (domain.Task, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, title, description, status, memo, project_id, parent_id, due_date,
		       tags, priority_json, created_at, updated_at, completed_at
		FROM tasks WHERE id = ?`, string(id))
	return scanTask(row)
}

func (r *TaskRepository) FindActive(ctx context.Context) ([]domain.Task, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title, description, status, memo, project_id, parent_id, due_date,
		       tags, priority_json, created_at, updated_at, completed_at
		FROM tasks
		WHERE status IN ('todo','in_progress','waiting')
		ORDER BY priority_json DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTasks(rows)
}

func (r *TaskRepository) List(ctx context.Context, f ports.TaskFilter) ([]domain.Task, int, error) {
	// Build WHERE clause dynamically.
	query := `
		SELECT id, title, description, status, memo, project_id, parent_id, due_date,
		       tags, priority_json, created_at, updated_at, completed_at
		FROM tasks WHERE 1=1`
	args := []any{}

	if len(f.Statuses) > 0 {
		query += " AND status IN (?" + repeatedPlaceholders(len(f.Statuses)-1) + ")"
		for _, s := range f.Statuses {
			args = append(args, string(s))
		}
	}
	if f.ProjectID != nil {
		query += " AND project_id = ?"
		args = append(args, *f.ProjectID)
	}

	countQuery := "SELECT COUNT(*) FROM (" + query + ")"
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	if f.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d OFFSET %d", f.Limit, f.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	tasks, err := scanTasks(rows)
	return tasks, total, err
}

func (r *TaskRepository) Delete(ctx context.Context, id domain.TaskID) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM tasks WHERE id = ?", string(id))
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("task_repository: delete %s: %w", id, domain.ErrNotFound)
	}
	return nil
}

func (r *TaskRepository) NextPriority(ctx context.Context) (*domain.Task, error) {
	tasks, err := r.FindActive(ctx)
	if err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		return nil, nil
	}
	// FindActive already returns sorted tasks; first is highest priority.
	return &tasks[0], nil
}

// ---- helpers ----

type scanner interface {
	Scan(dest ...any) error
}

func scanTask(s scanner) (domain.Task, error) {
	var (
		t           domain.Task
		status      string
		dueDate     *string
		completedAt *string
		parentID    *string
		tagsJSON    string
		priorityJSON string
		createdAt   string
		updatedAt   string
	)
	err := s.Scan(
		&t.ID, &t.Title, &t.Description, &status, &t.Memo,
		&t.ProjectID, &parentID, &dueDate,
		&tagsJSON, &priorityJSON,
		&createdAt, &updatedAt, &completedAt,
	)
	if err == sql.ErrNoRows {
		return domain.Task{}, fmt.Errorf("%w", domain.ErrNotFound)
	}
	if err != nil {
		return domain.Task{}, err
	}

	t.Status = domain.TaskStatus(status)

	if parentID != nil {
		pid := domain.TaskID(*parentID)
		t.ParentID = &pid
	}
	if dueDate != nil {
		d, _ := time.Parse(time.RFC3339, *dueDate)
		t.DueDate = &d
	}
	if completedAt != nil {
		c, _ := time.Parse(time.RFC3339, *completedAt)
		t.CompletedAt = &c
	}
	t.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	t.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	_ = json.Unmarshal([]byte(tagsJSON), &t.Tags)
	_ = json.Unmarshal([]byte(priorityJSON), &t.Priority)

	return t, nil
}

func scanTasks(rows *sql.Rows) ([]domain.Task, error) {
	var tasks []domain.Task
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func repeatedPlaceholders(n int) string {
	s := ""
	for i := 0; i < n; i++ {
		s += ",?"
	}
	return s
}
