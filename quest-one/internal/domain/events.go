package domain

import "time"

// DomainEvent is a marker interface for all domain events.
type DomainEvent interface {
	domainEvent()
}

// TaskCreatedEvent is emitted when a new task is persisted.
type TaskCreatedEvent struct {
	TaskID    TaskID
	CreatedAt time.Time
}

func (TaskCreatedEvent) domainEvent() {}

// TaskCompletedEvent is emitted when a task transitions to Done.
type TaskCompletedEvent struct {
	TaskID      TaskID
	CompletedAt time.Time
}

func (TaskCompletedEvent) domainEvent() {}

// TaskCancelledEvent is emitted when a task is cancelled.
type TaskCancelledEvent struct {
	TaskID TaskID
}

func (TaskCancelledEvent) domainEvent() {}

// TaskPriorityChangedEvent is emitted when a task's priority score is updated.
type TaskPriorityChangedEvent struct {
	TaskID      TaskID
	OldPriority PriorityScore
	NewPriority PriorityScore
}

func (TaskPriorityChangedEvent) domainEvent() {}

// TaskUpdatedEvent is emitted when task metadata (title, description, etc.) changes.
type TaskUpdatedEvent struct {
	TaskID    TaskID
	UpdatedAt time.Time
}

func (TaskUpdatedEvent) domainEvent() {}

// SourceItemSyncedEvent is emitted when a source item is fetched and persisted.
type SourceItemSyncedEvent struct {
	SourceItemID SourceItemID
	SourceType   SourceType
	IsNew        bool
}

func (SourceItemSyncedEvent) domainEvent() {}

// IntegrationConnectedEvent is emitted when a new integration is configured.
type IntegrationConnectedEvent struct {
	IntegrationID IntegrationID
	Provider      IntegrationProvider
}

func (IntegrationConnectedEvent) domainEvent() {}
