package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/HiteshRepo/awesome-tools/atlassian"
)

func fetchJiraTickets(ctx context.Context, from, to string) ([]atlassian.TicketSummary, error) {
	cfg := atlassian.Config{
		// Preferred: reuse OAuth token stored by Claude Code (no extra credentials needed).
		ClaudeCodeMCP: atlassian.ClaudeCodeMCPConfig{
			ServerName: "atlassian-vdc-workspace",
			CloudID:    cloudID(),
		},
		// Fallbacks for environments where Claude Code MCP is unavailable.
		JiraREST: atlassian.JiraRESTConfig{
			BaseURL:  os.Getenv("JIRA_URL"),
			Email:    os.Getenv("JIRA_EMAIL"),
			APIToken: os.Getenv("JIRA_API_TOKEN"),
		},
		Jira: atlassian.ServerConfig{
			Command: os.Getenv("JIRA_COMMAND"),
			Args:    strings.Fields(os.Getenv("JIRA_ARGS")),
		},
	}

	client, err := atlassian.NewClient(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("atlassian client: %w", err)
	}
	defer client.Close()

	jql := fmt.Sprintf(`assignee = currentUser() AND updated >= "%s" AND updated <= "%s" ORDER BY updated DESC`, from, to)
	result, err := client.SearchTickets(ctx, jql, 50)
	if err != nil {
		return nil, fmt.Errorf("search tickets: %w", err)
	}

	return result.Tickets, nil
}

// cloudID returns the Atlassian cloud ID to pass to MCP tools.
// Reads JIRA_CLOUD_ID env var; defaults to the Veeam workspace.
func cloudID() string {
	if id := os.Getenv("JIRA_CLOUD_ID"); id != "" {
		return id
	}
	return "veeam-vdc.atlassian.net"
}
