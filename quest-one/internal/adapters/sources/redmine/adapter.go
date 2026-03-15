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
	"strconv"
	"sync"
	"time"

	"github.com/quest-one/quest-one/internal/adapters/keychain"
	"github.com/quest-one/quest-one/internal/domain"
)

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
	account := keychain.AccountKey(string(a.SourceType()), string(integration.ID))
	apiKey, err := a.secrets.Get(keychain.ServiceName, account)
	if err != nil {
		return nil, fmt.Errorf("redmine adapter: credentials: %w", err)
	}

	limit := 100
	if integration.SyncFilters.MaxItems > 0 {
		limit = integration.SyncFilters.MaxItems
	}

	// Determine which project keys to query; empty means all assigned issues.
	projectKeys := integration.SyncFilters.ProjectKeys
	if len(projectKeys) == 0 {
		projectKeys = []string{""}
	}

	now := time.Now().UTC()
	batches := make([][]redmineIssue, len(projectKeys))
	errs := make([]error, len(projectKeys))
	var wg sync.WaitGroup
	for i, key := range projectKeys {
		wg.Add(1)
		go func(idx int, projectKey string) {
			defer wg.Done()
			batches[idx], errs[idx] = a.fetchIssues(ctx, integration.BaseURL, apiKey, projectKey, limit)
		}(i, key)
	}
	wg.Wait()

	var items []domain.SourceItem
	for i, batch := range batches {
		if errs[i] != nil {
			return nil, errs[i]
		}
		for _, issue := range batch {
			item := domain.NewSourceItem(
				domain.SourceTypeRedmine,
				strconv.Itoa(issue.ID),
				issue.Subject,
			)
			item.Description = issue.Description
			item.URL = fmt.Sprintf("%s/issues/%d", integration.BaseURL, issue.ID)
			item.Priority = issue.Priority.ID
			item.Status = issue.Status.Name
			item.ProjectID = issue.Project.Identifier
			item.AssigneeID = strconv.Itoa(issue.AssignedTo.ID)
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
	}
	return items, nil
}

func (a *Adapter) fetchIssues(ctx context.Context, baseURL, apiKey, projectKey string, limit int) ([]redmineIssue, error) {
	params := url.Values{
		"assigned_to_id": {"me"},
		"status_id":      {"open"},
		"limit":          {strconv.Itoa(limit)},
	}
	if projectKey != "" {
		params.Set("project_id", projectKey)
	}

	reqURL := fmt.Sprintf("%s/issues.json?%s", baseURL, params.Encode())
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
	return body.Issues, nil
}

type redmineIssue struct {
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
}

type redmineIssuesResponse struct {
	Issues []redmineIssue `json:"issues"`
}
