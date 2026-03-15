// Package priority provides the heuristic priority engine.
package priority

import (
	"context"
	"time"

	"github.com/quest-one/quest-one/internal/domain"
	"github.com/quest-one/quest-one/internal/ports"
)

// HeuristicEngine recalculates priority scores using rule-based heuristics
// without any external API calls.
type HeuristicEngine struct{}

var _ ports.PriorityEngine = (*HeuristicEngine)(nil)

// New returns a HeuristicEngine.
func New() *HeuristicEngine {
	return &HeuristicEngine{}
}

// Recalculate updates DeadlineUrgency and WaitingDays for each task.
func (e *HeuristicEngine) Recalculate(_ context.Context, tasks []domain.Task) ([]domain.Task, error) {
	now := time.Now().UTC()
	for i := range tasks {
		t := &tasks[i]

		// Deadline urgency
		if t.DueDate != nil {
			t.Priority.DeadlineUrgency = domain.DeadlineUrgencyScore(domain.DaysUntil(*t.DueDate))
		}

		// Waiting days
		t.Priority.WaitingDays = int(now.Sub(t.CreatedAt).Hours() / 24)

		// CreationOrder is the ULID string, which sorts lexicographically by time.
		if t.Priority.CreationOrder == "" {
			t.Priority.CreationOrder = string(t.ID)
		}
	}
	return tasks, nil
}
