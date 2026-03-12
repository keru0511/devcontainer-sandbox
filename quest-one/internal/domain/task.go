package domain

import (
	"time"
)

// TaskStatus represents the lifecycle state of a task.
type TaskStatus string

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusDone       TaskStatus = "done"
	TaskStatusCancelled  TaskStatus = "cancelled"
	TaskStatusWaiting    TaskStatus = "waiting"
)

// TaskID is a ULID string identifier for a task.
type TaskID string

// Task is the core aggregate root.
type Task struct {
	ID          TaskID
	Title       string
	Description string
	Status      TaskStatus
	Priority    PriorityScore
	DueDate     *time.Time
	SourceItems []SourceItemRef
	Tags        []string
	ProjectID   *string
	ParentID    *TaskID
	Memo        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CompletedAt *time.Time
}

// SourceItemRef links a task to an external source item.
type SourceItemRef struct {
	SourceItemID SourceItemID
	SourceType   SourceType
}

// NewTask constructs a new task with the given title and a default priority.
func NewTask(id TaskID, title string) Task {
	now := time.Now().UTC()
	return Task{
		ID:        id,
		Title:     title,
		Status:    TaskStatusTodo,
		Priority:  DefaultPriorityScore(),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// IsActive returns true if the task is not yet finished.
func (t *Task) IsActive() bool {
	return t.Status == TaskStatusTodo || t.Status == TaskStatusInProgress || t.Status == TaskStatusWaiting
}

// Complete marks the task as done and records the completion time.
func (t *Task) Complete() DomainEvent {
	now := time.Now().UTC()
	t.Status = TaskStatusDone
	t.CompletedAt = &now
	t.UpdatedAt = now
	return TaskCompletedEvent{TaskID: t.ID, CompletedAt: now}
}

// Cancel marks the task as cancelled.
func (t *Task) Cancel() DomainEvent {
	now := time.Now().UTC()
	t.Status = TaskStatusCancelled
	t.UpdatedAt = now
	return TaskCancelledEvent{TaskID: t.ID}
}

// UpdatePriority replaces the task's priority score.
func (t *Task) UpdatePriority(score PriorityScore) DomainEvent {
	old := t.Priority
	t.Priority = score
	t.UpdatedAt = time.Now().UTC()
	return TaskPriorityChangedEvent{TaskID: t.ID, OldPriority: old, NewPriority: score}
}

// AddMemo appends or replaces the memo text.
func (t *Task) AddMemo(text string) {
	t.Memo = text
	t.UpdatedAt = time.Now().UTC()
}

// Promote increases the manual urgency of a task by one level.
func (t *Task) Promote() DomainEvent {
	if t.Priority.ManualUrgency < UrgencyMax {
		t.Priority.ManualUrgency++
		t.UpdatedAt = time.Now().UTC()
	}
	return TaskPriorityChangedEvent{TaskID: t.ID, OldPriority: t.Priority, NewPriority: t.Priority}
}
