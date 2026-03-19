package atlassian

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
)

// ErrACLINotFound is returned when the acli binary is not on PATH.
var ErrACLINotFound = errors.New("acli binary not found")

func (c *Client) acliSearch(ctx context.Context, jql string, limit int) (*SearchResult, error) {
	path, err := exec.LookPath("acli")
	if err != nil {
		return nil, ErrACLINotFound
	}

	args := []string{
		"jira", "workitem", "search",
		"--jql", jql,
		"--json",
		"--limit", strconv.Itoa(limit),
	}

	out, err := exec.CommandContext(ctx, path, args...).Output()
	if err != nil {
		return nil, fmt.Errorf("acli search: %w", err)
	}

	return parseACLISearchResult(out)
}

func parseACLISearchResult(data []byte) (*SearchResult, error) {
	var issues []struct {
		Key     string `json:"key"`
		Summary string `json:"summary"`
		Status  struct {
			Name string `json:"name"`
		} `json:"status"`
		Priority struct {
			Name string `json:"name"`
		} `json:"priority"`
		Assignee struct {
			DisplayName string `json:"displayName"`
		} `json:"assignee"`
	}

	if err := json.Unmarshal(data, &issues); err != nil {
		return nil, fmt.Errorf("parse acli result: %w", err)
	}

	result := &SearchResult{Total: len(issues)}
	for _, issue := range issues {
		result.Tickets = append(result.Tickets, TicketSummary{
			ID:       issue.Key,
			Summary:  issue.Summary,
			Status:   issue.Status.Name,
			Priority: issue.Priority.Name,
			Assignee: issue.Assignee.DisplayName,
		})
	}
	return result, nil
}
