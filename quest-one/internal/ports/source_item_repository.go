package ports

import (
	"context"

	"github.com/quest-one/quest-one/internal/domain"
)

// SourceItemRepository is the persistence port for SourceItem aggregates.
type SourceItemRepository interface {
	// Save inserts or updates a source item.
	Save(ctx context.Context, item domain.SourceItem) error

	// FindByID retrieves a source item by its ULID.
	FindByID(ctx context.Context, id domain.SourceItemID) (domain.SourceItem, error)

	// FindByExternalID looks up a source item by its source type + external ID pair.
	// Returns ErrNotFound if not present.
	FindByExternalID(ctx context.Context, sourceType domain.SourceType, externalID string) (domain.SourceItem, error)

	// FindBySourceType returns all items from a specific source.
	FindBySourceType(ctx context.Context, sourceType domain.SourceType) ([]domain.SourceItem, error)

	// Delete removes a source item.
	Delete(ctx context.Context, id domain.SourceItemID) error
}
