package domain

import "time"

// SourceType identifies which external system a source item came from.
type SourceType string

const (
	SourceTypeRedmine SourceType = "redmine"
	SourceTypeSlack   SourceType = "slack"
	SourceTypeGCal    SourceType = "gcal"
	SourceTypeNotePM  SourceType = "notepm"
	SourceTypeGDrive  SourceType = "gdrive"
	SourceTypeManual  SourceType = "manual"
)

// SourceItemID is a ULID string identifier for a source item.
type SourceItemID string

// SourceItem represents a work item fetched from an external system.
// Multiple source items can be linked to a single task.
type SourceItem struct {
	ID           SourceItemID
	SourceType   SourceType
	ExternalID   string // ID in the external system
	Title        string
	Description  string
	URL          string
	Priority     int    // raw priority from source system
	Status       string // status string as reported by source
	AssigneeID   string // external user identifier
	ProjectID    string // external project identifier
	Labels       []string
	DueDate      *time.Time
	LastSyncedAt time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	RawPayload   []byte // JSON snapshot from last sync
}

// NewSourceItem constructs a new SourceItem.
func NewSourceItem(id SourceItemID, sourceType SourceType, externalID, title string) SourceItem {
	now := time.Now().UTC()
	return SourceItem{
		ID:           id,
		SourceType:   sourceType,
		ExternalID:   externalID,
		Title:        title,
		LastSyncedAt: now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// SourceKey returns a unique string combining source type and external ID.
func (s SourceItem) SourceKey() string {
	return string(s.SourceType) + ":" + s.ExternalID
}
