package application

import (
	"context"
	"fmt"

	"github.com/oklog/ulid/v2"
	"github.com/quest-one/quest-one/internal/domain"
)

// AddIntegrationInput is the DTO for registering a new integration.
type AddIntegrationInput struct {
	Provider domain.IntegrationProvider
	Name     string
	BaseURL  string
	Filters  domain.SyncFilters
}

// AddIntegration registers a new (disabled) integration.
func (a *App) AddIntegration(ctx context.Context, in AddIntegrationInput) (domain.Integration, error) {
	if in.Name == "" {
		return domain.Integration{}, fmt.Errorf("%w: name is required", domain.ErrInvalidInput)
	}

	id := domain.IntegrationID(ulid.Make().String())
	integration := domain.NewIntegration(id, in.Provider, in.Name, in.BaseURL)
	integration.SyncFilters = in.Filters

	if err := a.Integrations.Save(ctx, integration); err != nil {
		return domain.Integration{}, fmt.Errorf("add_integration: save: %w", err)
	}
	return integration, nil
}

// EnableIntegration activates an integration so it participates in sync.
func (a *App) EnableIntegration(ctx context.Context, id domain.IntegrationID) (domain.Integration, error) {
	integration, err := a.Integrations.FindByID(ctx, id)
	if err != nil {
		return domain.Integration{}, fmt.Errorf("enable_integration: find: %w", err)
	}

	event := integration.Enable()

	if err := a.Integrations.Save(ctx, integration); err != nil {
		return domain.Integration{}, fmt.Errorf("enable_integration: save: %w", err)
	}

	_ = a.Events.Publish(ctx, event)
	return integration, nil
}

// DisableIntegration deactivates an integration.
func (a *App) DisableIntegration(ctx context.Context, id domain.IntegrationID) (domain.Integration, error) {
	integration, err := a.Integrations.FindByID(ctx, id)
	if err != nil {
		return domain.Integration{}, fmt.Errorf("disable_integration: find: %w", err)
	}

	integration.Disable()

	if err := a.Integrations.Save(ctx, integration); err != nil {
		return domain.Integration{}, fmt.Errorf("disable_integration: save: %w", err)
	}
	return integration, nil
}

// ListIntegrations returns all registered integrations.
func (a *App) ListIntegrations(ctx context.Context) ([]domain.Integration, error) {
	return a.Integrations.FindAll(ctx)
}

// DeleteIntegration removes an integration configuration.
func (a *App) DeleteIntegration(ctx context.Context, id domain.IntegrationID) error {
	return a.Integrations.Delete(ctx, id)
}
