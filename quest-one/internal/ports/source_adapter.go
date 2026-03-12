package ports

import (
	"context"

	"github.com/quest-one/quest-one/internal/domain"
)

// SyncResult summarises what was fetched in a single sync run.
type SyncResult struct {
	SourceType domain.SourceType
	Created    int
	Updated    int
	Deleted    int
	Errors     []error
}

// SourceAdapter fetches items from an external system and converts them
// to domain.SourceItem values. Adapters are read-only with respect to the
// external system (AllowedClient guard).
type SourceAdapter interface {
	// SourceType returns the type of source this adapter handles.
	SourceType() domain.SourceType

	// Sync fetches the latest items from the external system.
	// The returned items are raw domain objects; callers persist them.
	Sync(ctx context.Context, integration domain.Integration) ([]domain.SourceItem, error)
}
