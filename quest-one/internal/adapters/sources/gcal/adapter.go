// Package gcal provides a stub source adapter for Google Calendar.
// Full OAuth2 support requires configuring a Google Cloud project and
// storing refresh tokens; this stub returns a clear error explaining what
// is needed.
package gcal

import (
	"context"
	"fmt"

	"github.com/quest-one/quest-one/internal/adapters/keychain"
	"github.com/quest-one/quest-one/internal/domain"
)

// Adapter is a stub for the Google Calendar integration.
// To activate, configure OAuth2 credentials and implement the full flow.
type Adapter struct {
	secrets keychain.SecretStore
}

// New creates a GCal adapter (stub).
func New(secrets keychain.SecretStore) *Adapter {
	return &Adapter{secrets: secrets}
}

// SourceType implements ports.SourceAdapter.
func (a *Adapter) SourceType() domain.SourceType { return domain.SourceTypeGCal }

// Sync returns an error explaining that GCal OAuth2 setup is required.
// implements ports.SourceAdapter.
func (a *Adapter) Sync(_ context.Context, integration domain.Integration) ([]domain.SourceItem, error) {
	return nil, fmt.Errorf(
		"gcal adapter (%s): OAuth2 credentials required — "+
			"create a Google Cloud project, enable the Calendar API, "+
			"download credentials.json, and run `quest-one integrations gcal-auth`",
		integration.Name,
	)
}
