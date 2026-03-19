package atlassian

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// ErrRESTNotConfigured is returned when JiraRESTConfig fields are missing.
var ErrRESTNotConfigured = errors.New("jira REST not configured")

func (c *Client) restSearch(ctx context.Context, jql string, limit int) (*SearchResult, error) {
	cfg := c.cfg.JiraREST
	if cfg.BaseURL == "" || cfg.Email == "" || cfg.APIToken == "" {
		return nil, ErrRESTNotConfigured
	}

	endpoint := cfg.BaseURL + "/rest/api/3/search"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("rest search request: %w", err)
	}

	q := url.Values{}
	q.Set("jql", jql)
	q.Set("maxResults", strconv.Itoa(limit))
	q.Set("fields", "summary,status,priority,assignee")
	req.URL.RawQuery = q.Encode()

	creds := base64.StdEncoding.EncodeToString([]byte(cfg.Email + ":" + cfg.APIToken))
	req.Header.Set("Authorization", "Basic "+creds)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("rest search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rest search: status %d", resp.StatusCode)
	}

	var payload struct {
		Total  int `json:"total"`
		Issues []struct {
			Key    string `json:"key"`
			Fields struct {
				Summary  string                            `json:"summary"`
				Status   struct{ Name string `json:"name"` } `json:"status"`
				Priority struct{ Name string `json:"name"` } `json:"priority"`
				Assignee struct{ DisplayName string `json:"displayName"` } `json:"assignee"`
			} `json:"fields"`
		} `json:"issues"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("parse rest result: %w", err)
	}

	result := &SearchResult{Total: payload.Total}
	for _, issue := range payload.Issues {
		result.Tickets = append(result.Tickets, TicketSummary{
			ID:       issue.Key,
			Summary:  issue.Fields.Summary,
			Status:   issue.Fields.Status.Name,
			Priority: issue.Fields.Priority.Name,
			Assignee: issue.Fields.Assignee.DisplayName,
		})
	}
	return result, nil
}
