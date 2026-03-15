// Package notepm implements a source adapter for NotePM.
// It fetches notes assigned to or mentioning the authenticated user.
//
// Credentials: store the API token in the OS keychain with
//
//	service = "quest-one"
//	account = "notepm.<integration_id>"
//
// or export QUEST_ONE_NOTEPM_<INTEGRATION_ID> as an env var.
package notepm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/quest-one/quest-one/internal/adapters/keychain"
	"github.com/quest-one/quest-one/internal/domain"
)

// Adapter fetches NotePM notes accessible to the authenticated user.
type Adapter struct {
	secrets keychain.SecretStore
	client  *http.Client
}

// New creates a NotePM adapter backed by the given secret store.
func New(secrets keychain.SecretStore) *Adapter {
	return &Adapter{
		secrets: secrets,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

// SourceType implements ports.SourceAdapter.
func (a *Adapter) SourceType() domain.SourceType { return domain.SourceTypeNotePM }

// Sync fetches recent NotePM notes as source items.
// implements ports.SourceAdapter.
func (a *Adapter) Sync(ctx context.Context, integration domain.Integration) ([]domain.SourceItem, error) {
	account := keychain.AccountKey(string(a.SourceType()), string(integration.ID))
	token, err := a.secrets.Get(keychain.ServiceName, account)
	if err != nil {
		return nil, fmt.Errorf("notepm adapter: credentials: %w", err)
	}

	limit := 100
	if integration.SyncFilters.MaxItems > 0 {
		limit = integration.SyncFilters.MaxItems
	}

	// NotePM API: GET https://<team>.notepm.jp/api/v1/notes
	q := url.Values{}
	q.Set("per_page", strconv.Itoa(limit))
	reqURL := fmt.Sprintf("%s/api/v1/notes?%s", integration.BaseURL, q.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("notepm adapter: build request: %w", err)
	}
	req.Header.Set("X-Api-Token", token)
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("notepm adapter: http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("notepm adapter: unexpected status %d", resp.StatusCode)
	}

	var body notePMNotesResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("notepm adapter: decode: %w", err)
	}

	now := time.Now().UTC()
	items := make([]domain.SourceItem, 0, len(body.Notes))
	for _, note := range body.Notes {
		item := domain.NewSourceItem(
			domain.SourceTypeNotePM,
			strconv.Itoa(note.ID),
			note.Title,
		)
		item.Description = note.Body
		item.URL = fmt.Sprintf("%s/notes/%d", integration.BaseURL, note.ID)
		item.Status = note.Status
		item.LastSyncedAt = now
		raw, _ := json.Marshal(note)
		item.RawPayload = raw
		items = append(items, item)
	}
	return items, nil
}

type notePMNotesResponse struct {
	Notes []struct {
		ID     int    `json:"id"`
		Title  string `json:"title"`
		Body   string `json:"body"`
		Status string `json:"status"`
	} `json:"notes"`
}
