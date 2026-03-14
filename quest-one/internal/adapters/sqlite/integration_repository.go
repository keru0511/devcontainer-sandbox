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

// IntegrationRepository implements ports.IntegrationRepository.
type IntegrationRepository struct {
	db  *sql.DB
	log *slog.Logger
}

var _ ports.IntegrationRepository = (*IntegrationRepository)(nil)

func (r *IntegrationRepository) Save(ctx context.Context, i domain.Integration) error {
	filtersJSON, _ := json.Marshal(i.SyncFilters)

	var lastSynced *string
	if i.LastSyncedAt != nil {
		s := i.LastSyncedAt.UTC().Format(time.RFC3339)
		lastSynced = &s
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO integrations
			(id, provider, name, base_url, enabled, sync_filters, last_synced_at, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?)
		ON CONFLICT(id) DO UPDATE SET
			provider      = excluded.provider,
			name          = excluded.name,
			base_url      = excluded.base_url,
			enabled       = excluded.enabled,
			sync_filters  = excluded.sync_filters,
			last_synced_at = excluded.last_synced_at,
			updated_at    = excluded.updated_at`,
		string(i.ID), string(i.Provider), i.Name, i.BaseURL,
		boolToInt(i.Enabled), string(filtersJSON), lastSynced,
		i.CreatedAt.UTC().Format(time.RFC3339),
		i.UpdatedAt.UTC().Format(time.RFC3339),
	)
	return err
}

func (r *IntegrationRepository) FindByID(ctx context.Context, id domain.IntegrationID) (domain.Integration, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, provider, name, base_url, enabled, sync_filters, last_synced_at, created_at, updated_at
		 FROM integrations WHERE id = ?`, string(id))
	return scanIntegration(row)
}

func (r *IntegrationRepository) FindAll(ctx context.Context) ([]domain.Integration, error) {
	return r.findWhere(ctx, "1=1")
}

func (r *IntegrationRepository) FindEnabled(ctx context.Context) ([]domain.Integration, error) {
	return r.findWhere(ctx, "enabled = 1")
}

func (r *IntegrationRepository) findWhere(ctx context.Context, where string) ([]domain.Integration, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, provider, name, base_url, enabled, sync_filters, last_synced_at, created_at, updated_at
		 FROM integrations WHERE `+where)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Integration
	for rows.Next() {
		i, err := scanIntegration(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, i)
	}
	return out, rows.Err()
}

func (r *IntegrationRepository) Delete(ctx context.Context, id domain.IntegrationID) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM integrations WHERE id = ?", string(id))
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("integration_repository: delete %s: %w", id, domain.ErrNotFound)
	}
	return nil
}

func scanIntegration(s scanner) (domain.Integration, error) {
	var (
		i           domain.Integration
		provider    string
		enabledInt  int
		filtersJSON string
		lastSynced  *string
		createdAt   string
		updatedAt   string
	)
	err := s.Scan(
		&i.ID, &provider, &i.Name, &i.BaseURL,
		&enabledInt, &filtersJSON, &lastSynced,
		&createdAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		return domain.Integration{}, fmt.Errorf("%w", domain.ErrNotFound)
	}
	if err != nil {
		return domain.Integration{}, err
	}

	i.Provider = domain.IntegrationProvider(provider)
	i.Enabled = enabledInt != 0
	_ = json.Unmarshal([]byte(filtersJSON), &i.SyncFilters)
	if lastSynced != nil {
		t, _ := time.Parse(time.RFC3339, *lastSynced)
		i.LastSyncedAt = &t
	}
	i.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	i.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	return i, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
