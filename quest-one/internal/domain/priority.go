package domain

import "time"

// DaysUntil returns the number of whole days from now until t (negative if past).
func DaysUntil(t time.Time) int {
	return int(time.Until(t).Hours() / 24)
}

// UrgencyLevel is a manual urgency override (0 = none, higher = more urgent).
type UrgencyLevel int

const (
	UrgencyNone     UrgencyLevel = 0
	UrgencyLow      UrgencyLevel = 1
	UrgencyMedium   UrgencyLevel = 2
	UrgencyHigh     UrgencyLevel = 3
	UrgencyCritical UrgencyLevel = 4
	UrgencyMax      UrgencyLevel = UrgencyCritical
)

// ImpactLevel represents the estimated business impact of completing a task.
type ImpactLevel int

const (
	ImpactNone   ImpactLevel = 0
	ImpactLow    ImpactLevel = 1
	ImpactMedium ImpactLevel = 2
	ImpactHigh   ImpactLevel = 3
)

// PriorityScore is a multi-dimensional priority descriptor.
// The 9 fields are compared lexicographically (in order) by PriorityComparator.
//
// Comparison order (descending = higher priority first):
//  1. ManualUrgency     — explicit human override (highest weight)
//  2. DeadlineUrgency   — derived from due date proximity
//  3. BlocksCount       — tasks blocked by this one
//  4. Impact            — estimated business value
//  5. SourcePriority    — priority as reported by source system
//  6. WaitingDays       — how long the task has been waiting
//  7. DependencyDepth   — chain length of downstream tasks
//  8. RecencyScore      — recently interacted tasks get a boost
//  9. CreationOrder     — older tasks win ties (ULID sort)
type PriorityScore struct {
	ManualUrgency   UrgencyLevel
	DeadlineUrgency int // 0-100, derived from due date
	BlocksCount     int
	Impact          ImpactLevel
	SourcePriority  int // raw priority from source system
	WaitingDays     int
	DependencyDepth int
	RecencyScore    int    // 0-100
	CreationOrder   string // ULID of the task (lexicographically sortable)
}

// DefaultPriorityScore returns a zero-value priority suitable for new tasks.
func DefaultPriorityScore() PriorityScore {
	return PriorityScore{}
}

// DeadlineUrgencyScore computes a 0-100 urgency score from days until due date.
// 0 = no due date or far future. 100 = overdue.
func DeadlineUrgencyScore(daysUntilDue int) int {
	if daysUntilDue <= 0 {
		return 100
	}
	if daysUntilDue >= 30 {
		return 0
	}
	// Linear interpolation: 1 day = 97, 30 days = 0
	return 100 - (daysUntilDue * 100 / 30)
}
