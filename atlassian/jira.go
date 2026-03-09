package atlassian

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// Ticket holds the full details of a Jira issue.
type Ticket struct {
	ID                 string
	URL                string
	Summary            string
	Description        string
	Status             string
	Priority           string
	Labels             []string
	Components         []string
	ConfluenceLinks    []string
	LinkedPRs          []string
	AcceptanceCriteria []string
	Comments           []Comment
}

// Comment is a single comment on a Jira issue.
type Comment struct {
	Author    string
	Body      string
	CreatedAt string
}

// TicketSummary is a lightweight representation used in search results.
type TicketSummary struct {
	ID       string
	Summary  string
	Status   string
	Priority string
	Assignee string
}

// SearchResult contains the results of a JQL search.
type SearchResult struct {
	Total   int
	Tickets []TicketSummary
}

// GetTicket fetches the full details of a Jira issue by its key (e.g. "PROJ-123").
func (c *Client) GetTicket(ctx context.Context, ticketID string) (*Ticket, error) {
	req := mcp.CallToolRequest{}
	req.Params.Name = "jira_get_issue"
	req.Params.Arguments = map[string]any{
		"issue_key":     ticketID,
		"fields":        "*all",
		"comment_limit": 10,
	}

	resp, err := c.jira.CallTool(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("jira_get_issue: %w", err)
	}

	return parseJiraIssue(ticketID, extractText(resp))
}

// SearchTickets runs a JQL query and returns up to limit results.
func (c *Client) SearchTickets(ctx context.Context, jql string, limit int) (*SearchResult, error) {
	req := mcp.CallToolRequest{}
	req.Params.Name = "jira_search"
	req.Params.Arguments = map[string]any{
		"jql":    jql,
		"fields": "summary,status,priority,assignee",
		"limit":  limit,
	}

	resp, err := c.jira.CallTool(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("jira_search: %w", err)
	}

	return parseSearchResult(extractText(resp))
}

func parseJiraIssue(id, raw string) (*Ticket, error) {
	var payload struct {
		Key         string `json:"key"`
		URL         string `json:"url"`
		Summary     string `json:"summary"`
		Description string `json:"description"`
		Status      struct {
			Name string `json:"name"`
		} `json:"status"`
		Priority struct {
			Name string `json:"name"`
		} `json:"priority"`
		Labels     []string `json:"labels"`
		Components []string `json:"components"`
		Comments   []struct {
			Author struct {
				DisplayName string `json:"display_name"`
			} `json:"author"`
			Body    string `json:"body"`
			Created string `json:"created"`
		} `json:"comments"`
	}

	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return &Ticket{
			ID:          id,
			Summary:     id,
			Description: raw,
		}, nil
	}

	t := &Ticket{
		ID:          payload.Key,
		URL:         payload.URL,
		Summary:     payload.Summary,
		Description: payload.Description,
		Status:      payload.Status.Name,
		Priority:    payload.Priority.Name,
		Labels:      payload.Labels,
		Components:  payload.Components,
	}

	for _, c := range payload.Comments {
		t.Comments = append(t.Comments, Comment{
			Author:    c.Author.DisplayName,
			Body:      c.Body,
			CreatedAt: c.Created,
		})
	}
	if len(t.Comments) > 5 {
		t.Comments = t.Comments[len(t.Comments)-5:]
	}

	return t, nil
}

func parseSearchResult(raw string) (*SearchResult, error) {
	var payload struct {
		Total  int `json:"total"`
		Issues []struct {
			Key      string                         `json:"key"`
			Summary  string                         `json:"summary"`
			Status   struct{ Name string `json:"name"` } `json:"status"`
			Priority struct{ Name string `json:"name"` } `json:"priority"`
			Assignee struct{ DisplayName string `json:"display_name"` } `json:"assignee"`
		} `json:"issues"`
	}

	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, fmt.Errorf("parse search result: %w", err)
	}

	result := &SearchResult{Total: payload.Total}
	for _, issue := range payload.Issues {
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
