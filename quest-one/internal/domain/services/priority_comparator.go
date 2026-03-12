// Package services contains pure domain services with no external dependencies.
package services

import (
	"github.com/quest-one/quest-one/internal/domain"
)

// PriorityComparator compares two tasks using 9-step lexicographic comparison.
// Returns negative if a has higher priority than b, positive if b wins, 0 if equal.
func Compare(a, b domain.PriorityScore) int {
	// Step 1: ManualUrgency (higher = more urgent)
	if diff := int(b.ManualUrgency) - int(a.ManualUrgency); diff != 0 {
		return diff
	}

	// Step 2: DeadlineUrgency (higher = more urgent)
	if diff := b.DeadlineUrgency - a.DeadlineUrgency; diff != 0 {
		return diff
	}

	// Step 3: BlocksCount (more blocks = higher priority)
	if diff := b.BlocksCount - a.BlocksCount; diff != 0 {
		return diff
	}

	// Step 4: Impact (higher = more important)
	if diff := int(b.Impact) - int(a.Impact); diff != 0 {
		return diff
	}

	// Step 5: SourcePriority (higher raw priority = more urgent)
	if diff := b.SourcePriority - a.SourcePriority; diff != 0 {
		return diff
	}

	// Step 6: WaitingDays (longer wait = should do sooner)
	if diff := b.WaitingDays - a.WaitingDays; diff != 0 {
		return diff
	}

	// Step 7: DependencyDepth (deeper chain = do sooner to unblock)
	if diff := b.DependencyDepth - a.DependencyDepth; diff != 0 {
		return diff
	}

	// Step 8: RecencyScore (recently touched = higher priority)
	if diff := b.RecencyScore - a.RecencyScore; diff != 0 {
		return diff
	}

	// Step 9: CreationOrder — older tasks (lower ULID) win ties.
	// ULID is lexicographically sortable by creation time, so older < newer.
	if a.CreationOrder < b.CreationOrder {
		return -1
	}
	if a.CreationOrder > b.CreationOrder {
		return 1
	}
	return 0
}

// SortByPriority sorts a slice of tasks in-place by priority (highest first).
// Uses a stable sort to preserve insertion order for equal priorities.
func SortByPriority(tasks []domain.Task) {
	n := len(tasks)
	// Insertion sort — stable and fast for typical task list sizes (<1000).
	for i := 1; i < n; i++ {
		for j := i; j > 0 && Compare(tasks[j].Priority, tasks[j-1].Priority) < 0; j-- {
			tasks[j], tasks[j-1] = tasks[j-1], tasks[j]
		}
	}
}

// Top returns the highest-priority task from a non-empty slice.
// Returns nil if the slice is empty.
func Top(tasks []domain.Task) *domain.Task {
	if len(tasks) == 0 {
		return nil
	}
	best := &tasks[0]
	for i := 1; i < len(tasks); i++ {
		if Compare(tasks[i].Priority, best.Priority) < 0 {
			best = &tasks[i]
		}
	}
	return best
}
