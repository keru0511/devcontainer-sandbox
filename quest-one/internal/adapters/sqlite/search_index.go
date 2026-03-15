package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/quest-one/quest-one/internal/domain"
	"github.com/quest-one/quest-one/internal/ports"
)

// SearchIndex implements ports.SearchIndex using SQLite FTS5.
// The FTS table is kept in sync automatically via triggers defined in the schema.
type SearchIndex struct {
	db  *sql.DB
	log *slog.Logger
}

var _ ports.SearchIndex = (*SearchIndex)(nil)

func (s *SearchIndex) Index(_ context.Context, _ domain.Task) error {
	// No-op: FTS triggers handle indexing automatically on INSERT/UPDATE.
	return nil
}

func (s *SearchIndex) Remove(_ context.Context, _ domain.TaskID) error {
	// No-op: FTS triggers handle removal automatically on DELETE.
	return nil
}

func (s *SearchIndex) Search(ctx context.Context, query string, limit int) ([]ports.SearchResult, error) {
	if limit <= 0 {
		limit = 20
	}

	// bm25 returns negative values; ORDER BY ASC gives best (most negative) first.
	rows, err := s.db.QueryContext(ctx, `
		SELECT t.id, t.title, t.description, t.status, t.memo, t.project_id, t.parent_id,
		       t.due_date, t.tags, t.priority_json, t.created_at, t.updated_at, t.completed_at,
		       bm25(tasks_fts) AS score
		FROM tasks_fts
		JOIN tasks t ON t.id = tasks_fts.id
		WHERE tasks_fts MATCH ?
		ORDER BY score
		LIMIT ?`, query, limit)
	if err != nil {
		return nil, fmt.Errorf("search_index: query: %w", err)
	}
	defer rows.Close()

	var results []ports.SearchResult
	for rows.Next() {
		var (
			t            domain.Task
			status       string
			dueDate      *string
			completedAt  *string
			parentID     *string
			tagsJSON     string
			priorityJSON string
			createdAt    string
			updatedAt    string
			score        float64
		)
		err := rows.Scan(
			&t.ID, &t.Title, &t.Description, &status, &t.Memo,
			&t.ProjectID, &parentID, &dueDate,
			&tagsJSON, &priorityJSON,
			&createdAt, &updatedAt, &completedAt,
			&score,
		)
		if err != nil {
			return nil, fmt.Errorf("search_index: scan: %w", err)
		}

		t.Status = domain.TaskStatus(status)
		if parentID != nil {
			pid := domain.TaskID(*parentID)
			t.ParentID = &pid
		}
		t.DueDate = parseNullableTime(dueDate)
		t.CompletedAt = parseNullableTime(completedAt)
		t.CreatedAt = parseTime(createdAt)
		t.UpdatedAt = parseTime(updatedAt)
		_ = json.Unmarshal([]byte(tagsJSON), &t.Tags)
		_ = json.Unmarshal([]byte(priorityJSON), &t.Priority)

		// bm25 is negative; negate for a positive relevance score.
		results = append(results, ports.SearchResult{Task: t, Score: -score})
	}
	return results, rows.Err()
}
