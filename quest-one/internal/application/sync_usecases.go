package application

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/oklog/ulid/v2"
	"github.com/quest-one/quest-one/internal/domain"
	"github.com/quest-one/quest-one/internal/ports"
)

// SyncAll runs all enabled integrations' source adapters in sequence.
// Each adapter is given its own context-derived sub-span for observability.
func (a *App) SyncAll(ctx context.Context, adapters []ports.SourceAdapter) ([]ports.SyncResult, error) {
	integrations, err := a.Integrations.FindEnabled(ctx)
	if err != nil {
		return nil, fmt.Errorf("sync_all: find_enabled: %w", err)
	}

	adapterMap := make(map[domain.SourceType]ports.SourceAdapter, len(adapters))
	for _, ad := range adapters {
		adapterMap[ad.SourceType()] = ad
	}

	var results []ports.SyncResult
	for _, integration := range integrations {
		ad, ok := adapterMap[domain.SourceType(integration.Provider)]
		if !ok {
			a.Log.Warn("no adapter for integration", slog.String("provider", string(integration.Provider)))
			continue
		}
		r, err := a.syncIntegration(ctx, integration, ad)
		if err != nil {
			a.Log.Error("sync failed", slog.String("integration", string(integration.ID)), slog.Any("err", err))
		}
		results = append(results, r)
	}
	return results, nil
}

func (a *App) syncIntegration(
	ctx context.Context,
	integration domain.Integration,
	adapter ports.SourceAdapter,
) (ports.SyncResult, error) {
	result := ports.SyncResult{SourceType: adapter.SourceType()}

	items, err := adapter.Sync(ctx, integration)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("fetch: %w", err))
		return result, err
	}

	for _, item := range items {
		existing, findErr := a.SourceItems.FindByExternalID(ctx, item.SourceType, item.ExternalID)
		isNew := domain.IsNotFound(findErr)

		if !isNew {
			// Preserve the original ULID so we do an update, not a new insert.
			item.ID = existing.ID
			item.CreatedAt = existing.CreatedAt
			result.Updated++
		} else {
			item.ID = domain.SourceItemID(ulid.Make().String())
			result.Created++
		}

		if saveErr := a.SourceItems.Save(ctx, item); saveErr != nil {
			result.Errors = append(result.Errors, saveErr)
			continue
		}

		_ = a.Events.Publish(ctx, domain.SourceItemSyncedEvent{
			SourceItemID: item.ID,
			SourceType:   item.SourceType,
			IsNew:        isNew,
		})
	}

	return result, nil
}
