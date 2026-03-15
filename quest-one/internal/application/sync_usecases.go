package application

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/oklog/ulid/v2"

	"github.com/quest-one/quest-one/internal/domain"
	"github.com/quest-one/quest-one/internal/ports"
)

// SyncAll runs all enabled integrations' source adapters concurrently.
// If adapters is nil, the app's registered Adapters are used.
func (a *App) SyncAll(ctx context.Context, adapters []ports.SourceAdapter) ([]ports.SyncResult, error) {
	if adapters == nil {
		adapters = a.Adapters
	}
	integrations, err := a.Integrations.FindEnabled(ctx)
	if err != nil {
		return nil, fmt.Errorf("sync_all: find_enabled: %w", err)
	}

	adapterMap := make(map[domain.SourceType]ports.SourceAdapter, len(adapters))
	for _, ad := range adapters {
		adapterMap[ad.SourceType()] = ad
	}

	results := make([]ports.SyncResult, len(integrations))
	var wg sync.WaitGroup

	for i, integration := range integrations {
		ad, ok := adapterMap[domain.SourceType(integration.Provider)]
		if !ok {
			a.Log.Warn("no adapter for integration", slog.String("provider", string(integration.Provider)))
			continue
		}
		wg.Add(1)
		go func(idx int, intg domain.Integration, adapter ports.SourceAdapter) {
			defer wg.Done()
			r, err := a.syncIntegration(ctx, intg, adapter)
			if err != nil {
				a.Log.Error("sync failed", slog.String("integration", string(intg.ID)), slog.Any("err", err))
			}
			results[idx] = r // safe: each goroutine writes to a distinct index
		}(i, integration, ad)
	}
	wg.Wait()

	// Filter out zero-value results for skipped integrations.
	var out []ports.SyncResult
	for _, r := range results {
		if r.SourceType != "" {
			out = append(out, r)
		}
	}
	return out, nil
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

	// Batch-load existing source items once to avoid N+1 queries.
	existing, err := a.SourceItems.FindBySourceType(ctx, adapter.SourceType())
	if err != nil {
		return result, fmt.Errorf("syncIntegration: load_existing: %w", err)
	}
	existingByExternalID := make(map[string]domain.SourceItem, len(existing))
	for _, e := range existing {
		existingByExternalID[e.ExternalID] = e
	}

	for _, item := range items {
		isNew := true
		if prev, found := existingByExternalID[item.ExternalID]; found {
			// Preserve the original ULID so we do an update, not a new insert.
			item.ID = prev.ID
			item.CreatedAt = prev.CreatedAt
			result.Updated++
			isNew = false
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
