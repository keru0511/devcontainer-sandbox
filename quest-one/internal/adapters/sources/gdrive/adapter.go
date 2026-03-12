// Package gdrive provides a stub source adapter for Google Drive.
// Full OAuth2 support requires configuring a Google Cloud project;
// this stub returns a clear error explaining what is needed.
package gdrive

import (
	"context"
	"fmt"

	"github.com/quest-one/quest-one/internal/adapters/keychain"
	"github.com/quest-one/quest-one/internal/domain"
)

// Adapter is a stub for the Google Drive integration.
type Adapter struct {
	secrets keychain.SecretStore
}

// New creates a GDrive adapter (stub).
func New(secrets keychain.SecretStore) *Adapter {
	return &Adapter{secrets: secrets}
}

// SourceType implements ports.SourceAdapter.
func (a *Adapter) SourceType() domain.SourceType { return domain.SourceTypeGDrive }

// Sync returns an error explaining that GDrive OAuth2 setup is required.
// implements ports.SourceAdapter.
func (a *Adapter) Sync(_ context.Context, integration domain.Integration) ([]domain.SourceItem, error) {
	return nil, fmt.Errorf(
		"gdrive adapter (%s): OAuth2 credentials required — "+
			"create a Google Cloud project, enable the Drive API, "+
			"and run `quest-one integrations gdrive-auth`",
		integration.Name,
	)
}
