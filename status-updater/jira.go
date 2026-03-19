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
		Jira: atlassian.ServerConfig{
			Command: os.Getenv("JIRA_COMMAND"),
			Args:    strings.Fields(os.Getenv("JIRA_ARGS")),
		},
		JiraREST: atlassian.JiraRESTConfig{
			BaseURL:  os.Getenv("JIRA_URL"),
			Email:    os.Getenv("JIRA_EMAIL"),
			APIToken: os.Getenv("JIRA_API_TOKEN"),
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
