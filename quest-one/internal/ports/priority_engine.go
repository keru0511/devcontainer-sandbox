package ports

import (
	"context"

	"github.com/quest-one/quest-one/internal/domain"
)

// PriorityEngine recalculates priority scores for a batch of tasks.
// Implementations may use heuristics, LLM suggestions, or a combination.
type PriorityEngine interface {
	// Recalculate updates the PriorityScore for each task based on current context.
	// Returns the updated tasks (not persisted; callers must call TaskRepository.Save).
	Recalculate(ctx context.Context, tasks []domain.Task) ([]domain.Task, error)
}
