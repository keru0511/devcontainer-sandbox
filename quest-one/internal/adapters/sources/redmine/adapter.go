// Package redmine implements a source adapter for Redmine issue tracker.
// It fetches issues assigned to the authenticated user via the Redmine REST API.
//
// Credentials: store the API key in the OS keychain with
//
//	service = "quest-one"
//	account = "redmine.<integration_id>"
//
// or export QUEST_ONE_REDMINE_<INTEGRATION_ID> as an env var (non-macOS).
package redmine

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/quest-one/quest-one/internal/adapters/keychain"
	"github.com/quest-one/quest-one/internal/domain"
)

const keychainService = "quest-one"

// Adapter fetches Redmine issues assigned to the authenticated user.
type Adapter struct {
	secrets keychain.SecretStore
	client  *http.Client
}

// New creates a Redmine adapter backed by the given secret store.
func New(secrets keychain.SecretStore) *Adapter {
	return &Adapter{
		secrets: secrets,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

// SourceType implements ports.SourceAdapter.
func (a *Adapter) SourceType() domain.SourceType { return domain.SourceTypeRedmine }

// Sync fetches open issues assigned to the current user.
// implements ports.SourceAdapter.
func (a *Adapter) Sync(ctx context.Context, integration domain.Integration) ([]domain.SourceItem, error) {
	account := fmt.Sprintf("redmine.%s", integration.ID)
	apiKey, err := a.secrets.Get(keychainService, account)
	if err != nil {
		return nil, fmt.Errorf("redmine adapter: credentials: %w", err)
	}

	params := url.Values{
		"assigned_to_id": {"me"},
		"status_id":      {"open"},
		"limit":          {"100"},
	}
	if integration.SyncFilters.MaxItems > 0 {
		params.Set("limit", fmt.Sprintf("%d", integration.SyncFilters.MaxItems))
	}
	if len(integration.SyncFilters.ProjectKeys) > 0 {
		params.Set("project_id", integration.SyncFilters.ProjectKeys[0])
	}

	reqURL := fmt.Sprintf("%s/issues.json?%s", integration.BaseURL, params.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("redmine adapter: build request: %w", err)
	}
	req.Header.Set("X-Redmine-API-Key", apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("redmine adapter: http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("redmine adapter: unexpected status %d", resp.StatusCode)
	}

	var body redmineIssuesResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("redmine adapter: decode: %w", err)
	}

	now := time.Now().UTC()
	items := make([]domain.SourceItem, 0, len(body.Issues))
	for _, issue := range body.Issues {
		item := domain.NewSourceItem(
			"", // ID assigned by sync use case
			domain.SourceTypeRedmine,
			fmt.Sprintf("%d", issue.ID),
			issue.Subject,
		)
		item.Description = issue.Description
		item.URL = fmt.Sprintf("%s/issues/%d", integration.BaseURL, issue.ID)
		item.Priority = issue.Priority.ID
		item.Status = issue.Status.Name
		item.ProjectID = issue.Project.Identifier
		item.AssigneeID = fmt.Sprintf("%d", issue.AssignedTo.ID)
		item.LastSyncedAt = now
		if issue.DueDate != "" {
			if t, err := time.Parse("2006-01-02", issue.DueDate); err == nil {
				item.DueDate = &t
			}
		}
		raw, _ := json.Marshal(issue)
		item.RawPayload = raw
		items = append(items, item)
	}
	return items, nil
}

// redmineIssuesResponse mirrors the Redmine /issues.json shape.
type redmineIssuesResponse struct {
	Issues []struct {
		ID          int    `json:"id"`
		Subject     string `json:"subject"`
		Description string `json:"description"`
		Priority    struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"priority"`
		Status struct {
			Name string `json:"name"`
		} `json:"status"`
		Project struct {
			Identifier string `json:"identifier"`
		} `json:"project"`
		AssignedTo struct {
			ID int `json:"id"`
		} `json:"assigned_to"`
		DueDate string `json:"due_date"`
	} `json:"issues"`
}
