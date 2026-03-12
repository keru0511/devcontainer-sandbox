package domain

import "time"

// IntegrationProvider identifies the external system.
type IntegrationProvider string

const (
	IntegrationProviderRedmine IntegrationProvider = "redmine"
	IntegrationProviderSlack   IntegrationProvider = "slack"
	IntegrationProviderGCal    IntegrationProvider = "gcal"
	IntegrationProviderNotePM  IntegrationProvider = "notepm"
	IntegrationProviderGDrive  IntegrationProvider = "gdrive"
)

// IntegrationID is a ULID string identifier for an integration config.
type IntegrationID string

// Integration stores the connection configuration for an external system.
// Secrets (API keys, tokens) are stored in macOS Keychain, not here.
type Integration struct {
	ID           IntegrationID
	Provider     IntegrationProvider
	Name         string  // user-facing display name
	BaseURL      string  // e.g., "https://redmine.example.com"
	Enabled      bool
	SyncFilters  SyncFilters
	LastSyncedAt *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// SyncFilters controls what gets imported from the external system.
type SyncFilters struct {
	// ProjectKeys limits sync to specific project IDs/keys (empty = all).
	ProjectKeys []string

	// AssignedToMe if true only imports items assigned to the authenticated user.
	AssignedToMe bool

	// StatusFilter limits which statuses are imported (empty = all open).
	StatusFilter []string

	// MaxItems caps the number of items imported per sync run (0 = unlimited).
	MaxItems int
}

// NewIntegration constructs a new disabled integration.
func NewIntegration(id IntegrationID, provider IntegrationProvider, name, baseURL string) Integration {
	now := time.Now().UTC()
	return Integration{
		ID:        id,
		Provider:  provider,
		Name:      name,
		BaseURL:   baseURL,
		Enabled:   false,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Enable activates the integration.
func (i *Integration) Enable() DomainEvent {
	i.Enabled = true
	i.UpdatedAt = time.Now().UTC()
	return IntegrationConnectedEvent{IntegrationID: i.ID, Provider: i.Provider}
}

// Disable deactivates the integration without deleting it.
func (i *Integration) Disable() {
	i.Enabled = false
	i.UpdatedAt = time.Now().UTC()
}
