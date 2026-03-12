// Package ports defines the interfaces (hexagonal ports) that separate domain
// logic from infrastructure concerns. All adapters implement these interfaces.
package ports

import (
	"context"

	"github.com/quest-one/quest-one/internal/domain"
)

// TaskFilter specifies criteria for listing tasks.
type TaskFilter struct {
	Statuses  []domain.TaskStatus // empty = all statuses
	Tags      []string            // empty = all tags
	ProjectID *string
	Search    string // full-text search query
	Limit     int    // 0 = no limit
	Offset    int
}

// TaskRepository is the persistence port for Task aggregates.
type TaskRepository interface {
	// Save inserts or updates a task (upsert semantics).
	Save(ctx context.Context, task domain.Task) error

	// FindByID retrieves a single task by its ULID.
	// Returns ErrNotFound if the task does not exist.
	FindByID(ctx context.Context, id domain.TaskID) (domain.Task, error)

	// FindActive returns all non-terminal tasks ordered by priority score.
	FindActive(ctx context.Context) ([]domain.Task, error)

	// List returns tasks matching the given filter.
	List(ctx context.Context, filter TaskFilter) ([]domain.Task, int, error)

	// Delete permanently removes a task.
	Delete(ctx context.Context, id domain.TaskID) error

	// NextPriority returns the highest-priority active task, or nil if none exists.
	NextPriority(ctx context.Context) (*domain.Task, error)
}
