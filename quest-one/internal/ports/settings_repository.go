package ports

import (
	"context"

	"github.com/quest-one/quest-one/internal/domain"
)

// SettingsRepository loads and persists application settings.
type SettingsRepository interface {
	Load(ctx context.Context) (domain.Settings, error)
	Save(ctx context.Context, s domain.Settings) error
}
