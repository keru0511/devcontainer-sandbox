package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/quest-one/quest-one/internal/domain"
)

// GetSettings returns the current application settings.
func (a *App) GetSettings(ctx context.Context) (domain.Settings, error) {
	return a.Settings.Load(ctx)
}

// UpdateSettings validates and persists new settings.
func (a *App) UpdateSettings(ctx context.Context, s domain.Settings) error {
	errs := s.Validate()
	if len(errs) > 0 {
		return fmt.Errorf("%w: %s", domain.ErrInvalidInput, strings.Join(errs, "; "))
	}
	return a.Settings.Save(ctx, s)
}
