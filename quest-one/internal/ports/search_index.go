package ports

import (
	"context"

	"github.com/quest-one/quest-one/internal/domain"
)

// SearchResult is a ranked match from a full-text search.
type SearchResult struct {
	Task  domain.Task
	Score float64 // relevance score (higher = more relevant)
}

// SearchIndex provides full-text search over tasks.
// The default implementation uses SQLite FTS5.
type SearchIndex interface {
	// Index adds or updates a task document in the search index.
	Index(ctx context.Context, task domain.Task) error

	// Remove deletes a task document from the index.
	Remove(ctx context.Context, id domain.TaskID) error

	// Search returns tasks matching the query, ordered by relevance.
	Search(ctx context.Context, query string, limit int) ([]SearchResult, error)
}
