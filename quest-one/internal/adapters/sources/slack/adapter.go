// Package slack implements a source adapter for Slack.
// It fetches reminders set by the authenticated user, treating each
// reminder as a task-candidate.
//
// Credentials: store the user OAuth token in the OS keychain with
//
//	service = "quest-one"
//	account = "slack.<integration_id>"
//
// or export QUEST_ONE_SLACK_<INTEGRATION_ID> as an env var.
package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/quest-one/quest-one/internal/adapters/keychain"
	"github.com/quest-one/quest-one/internal/domain"
)

const keychainService = "quest-one"
const slackAPIBase = "https://slack.com/api"

// Adapter fetches Slack reminders for the authenticated user.
type Adapter struct {
	secrets keychain.SecretStore
	client  *http.Client
}

// New creates a Slack adapter backed by the given secret store.
func New(secrets keychain.SecretStore) *Adapter {
	return &Adapter{
		secrets: secrets,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

// SourceType implements ports.SourceAdapter.
func (a *Adapter) SourceType() domain.SourceType { return domain.SourceTypeSlack }

// Sync fetches incomplete Slack reminders as source items.
// implements ports.SourceAdapter.
func (a *Adapter) Sync(ctx context.Context, integration domain.Integration) ([]domain.SourceItem, error) {
	account := fmt.Sprintf("slack.%s", integration.ID)
	token, err := a.secrets.Get(keychainService, account)
	if err != nil {
		return nil, fmt.Errorf("slack adapter: credentials: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, slackAPIBase+"/reminders.list", nil)
	if err != nil {
		return nil, fmt.Errorf("slack adapter: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("slack adapter: http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("slack adapter: unexpected status %d", resp.StatusCode)
	}

	var body slackRemindersResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("slack adapter: decode: %w", err)
	}
	if !body.OK {
		return nil, fmt.Errorf("slack adapter: api error: %s", body.Error)
	}

	now := time.Now().UTC()
	items := make([]domain.SourceItem, 0, len(body.Reminders))
	for _, r := range body.Reminders {
		if r.Complete {
			continue
		}
		item := domain.NewSourceItem(
			"",
			domain.SourceTypeSlack,
			r.ID,
			r.Text,
		)
		item.Status = "incomplete"
		item.LastSyncedAt = now
		if r.Time > 0 {
			t := time.Unix(int64(r.Time), 0).UTC()
			item.DueDate = &t
		}
		raw, _ := json.Marshal(r)
		item.RawPayload = raw
		items = append(items, item)
	}
	return items, nil
}

type slackRemindersResponse struct {
	OK        bool   `json:"ok"`
	Error     string `json:"error,omitempty"`
	Reminders []struct {
		ID       string `json:"id"`
		Text     string `json:"text"`
		Complete bool   `json:"complete"`
		Time     int64  `json:"time"` // unix timestamp or 0
	} `json:"reminders"`
}
