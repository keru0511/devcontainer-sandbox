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

const settingsKey = "app"

// SettingsRepository implements ports.SettingsRepository using a single-row kv table.
type SettingsRepository struct {
	db  *sql.DB
	log *slog.Logger
}

var _ ports.SettingsRepository = (*SettingsRepository)(nil)

func (r *SettingsRepository) Load(ctx context.Context) (domain.Settings, error) {
	var value string
	err := r.db.QueryRowContext(ctx,
		`SELECT value FROM settings WHERE key = ?`, settingsKey).Scan(&value)
	if err == sql.ErrNoRows {
		return domain.DefaultSettings(), nil
	}
	if err != nil {
		return domain.Settings{}, fmt.Errorf("settings_repository: load: %w", err)
	}

	var s domain.Settings
	if err := json.Unmarshal([]byte(value), &s); err != nil {
		return domain.Settings{}, fmt.Errorf("settings_repository: unmarshal: %w", err)
	}
	return s, nil
}

func (r *SettingsRepository) Save(ctx context.Context, s domain.Settings) error {
	b, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("settings_repository: marshal: %w", err)
	}

	_, err = r.db.ExecContext(ctx,
		`INSERT INTO settings (key, value) VALUES (?,?)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		settingsKey, string(b),
	)
	return err
}
