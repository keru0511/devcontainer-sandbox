package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/quest-one/quest-one/internal/domain"
	"github.com/quest-one/quest-one/internal/ports"
)

// SourceItemRepository implements ports.SourceItemRepository.
type SourceItemRepository struct {
	db  *sql.DB
	log *slog.Logger
}

var _ ports.SourceItemRepository = (*SourceItemRepository)(nil)

func (r *SourceItemRepository) Save(ctx context.Context, item domain.SourceItem) error {
	labelsJSON, _ := json.Marshal(item.Labels)

	var dueDate *string
	if item.DueDate != nil {
		s := item.DueDate.UTC().Format(time.RFC3339)
		dueDate = &s
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO source_items
			(id, source_type, external_id, title, description, url, priority, status,
			 assignee_id, project_id, labels, due_date, last_synced_at, created_at, updated_at, raw_payload)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(id) DO UPDATE SET
			title         = excluded.title,
			description   = excluded.description,
			url           = excluded.url,
			priority      = excluded.priority,
			status        = excluded.status,
			assignee_id   = excluded.assignee_id,
			project_id    = excluded.project_id,
			labels        = excluded.labels,
			due_date      = excluded.due_date,
			last_synced_at = excluded.last_synced_at,
			updated_at    = excluded.updated_at,
			raw_payload   = excluded.raw_payload`,
		string(item.ID), string(item.SourceType), item.ExternalID,
		item.Title, item.Description, item.URL, item.Priority, item.Status,
		item.AssigneeID, item.ProjectID, string(labelsJSON), dueDate,
		item.LastSyncedAt.UTC().Format(time.RFC3339),
		item.CreatedAt.UTC().Format(time.RFC3339),
		item.UpdatedAt.UTC().Format(time.RFC3339),
		item.RawPayload,
	)
	return err
}

func (r *SourceItemRepository) FindByID(ctx context.Context, id domain.SourceItemID) (domain.SourceItem, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, source_type, external_id, title, description, url, priority, status,
		        assignee_id, project_id, labels, due_date, last_synced_at, created_at, updated_at, raw_payload
		 FROM source_items WHERE id = ?`, string(id))
	return scanSourceItem(row)
}

func (r *SourceItemRepository) FindByExternalID(ctx context.Context, sourceType domain.SourceType, externalID string) (domain.SourceItem, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, source_type, external_id, title, description, url, priority, status,
		        assignee_id, project_id, labels, due_date, last_synced_at, created_at, updated_at, raw_payload
		 FROM source_items WHERE source_type = ? AND external_id = ?`,
		string(sourceType), externalID)
	return scanSourceItem(row)
}

func (r *SourceItemRepository) FindBySourceType(ctx context.Context, sourceType domain.SourceType) ([]domain.SourceItem, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, source_type, external_id, title, description, url, priority, status,
		        assignee_id, project_id, labels, due_date, last_synced_at, created_at, updated_at, raw_payload
		 FROM source_items WHERE source_type = ?`, string(sourceType))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.SourceItem
	for rows.Next() {
		item, err := scanSourceItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SourceItemRepository) Delete(ctx context.Context, id domain.SourceItemID) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM source_items WHERE id = ?", string(id))
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("source_item_repository: delete %s: %w", id, domain.ErrNotFound)
	}
	return nil
}

func scanSourceItem(s scanner) (domain.SourceItem, error) {
	var (
		item       domain.SourceItem
		sourceType string
		labelsJSON string
		dueDate    *string
		lastSynced string
		createdAt  string
		updatedAt  string
	)
	err := s.Scan(
		&item.ID, &sourceType, &item.ExternalID, &item.Title, &item.Description,
		&item.URL, &item.Priority, &item.Status, &item.AssigneeID, &item.ProjectID,
		&labelsJSON, &dueDate, &lastSynced, &createdAt, &updatedAt, &item.RawPayload,
	)
	if err == sql.ErrNoRows {
		return domain.SourceItem{}, fmt.Errorf("%w", domain.ErrNotFound)
	}
	if err != nil {
		return domain.SourceItem{}, err
	}

	item.SourceType = domain.SourceType(sourceType)
	_ = json.Unmarshal([]byte(labelsJSON), &item.Labels)
	if dueDate != nil {
		d, _ := time.Parse(time.RFC3339, *dueDate)
		item.DueDate = &d
	}
	item.LastSyncedAt, _ = time.Parse(time.RFC3339, lastSynced)
	item.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	item.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	return item, nil
}
