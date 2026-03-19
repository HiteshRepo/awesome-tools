package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

const (
	anthropicAPI   = "https://api.anthropic.com/v1/messages"
	summarizeModel = "claude-haiku-4-5-20251001"
)

// generateSummary calls Claude Haiku to produce clean past-tense bullet points
// from the connected view. Falls back to raw titles if ANTHROPIC_API_KEY is unset.
func generateSummary(ctx context.Context, view ConnectedView) (string, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return fallbackSummary(view), nil
	}

	prompt := buildPrompt(view)

	body, _ := json.Marshal(map[string]any{
		"model":      summarizeModel,
		"max_tokens": 1024,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicAPI, bytes.NewReader(body))
	if err != nil {
		return fallbackSummary(view), nil
	}
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fallbackSummary(view), nil
	}
	defer resp.Body.Close()

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil || len(result.Content) == 0 {
		return fallbackSummary(view), nil
	}

	return strings.TrimSpace(result.Content[0].Text), nil
}

func buildPrompt(view ConnectedView) string {
	var sb strings.Builder

	sb.WriteString("You are writing a developer's weekly status update.\n")
	sb.WriteString("Convert the items below into concise past-tense bullet points in plain English.\n")
	sb.WriteString("Rules:\n")
	sb.WriteString("- No ticket IDs, no PR numbers, no repo names\n")
	sb.WriteString("- Combine closely related items into one bullet\n")
	sb.WriteString("- Keep each bullet under 15 words\n")
	sb.WriteString("- Output only the bullet lines, each starting with '- '\n\n")

	if len(view.Linked) > 0 || len(view.NoGH) > 0 {
		sb.WriteString("Jira tickets:\n")
		for _, lt := range view.Linked {
			fmt.Fprintf(&sb, "- %s (%s)\n", lt.Ticket.Summary, lt.Ticket.Status)
		}
		for _, t := range view.NoGH {
			fmt.Fprintf(&sb, "- %s (%s)\n", t.Summary, t.Status)
		}
		sb.WriteString("\n")
	}

	if len(view.Orphaned) > 0 {
		sb.WriteString("Additional GitHub PRs (no Jira ticket):\n")
		for _, p := range view.Orphaned {
			fmt.Fprintf(&sb, "- %s\n", p.PR.Title)
		}
	}

	return sb.String()
}

// fallbackSummary returns raw summaries as bullets when the API key is absent.
func fallbackSummary(view ConnectedView) string {
	seen := make(map[string]bool)
	var lines []string

	add := func(s string) {
		if s != "" && !seen[s] {
			seen[s] = true
			lines = append(lines, "- "+s)
		}
	}

	for _, lt := range view.Linked {
		add(lt.Ticket.Summary)
	}
	for _, t := range view.NoGH {
		add(t.Summary)
	}
	for _, p := range view.Orphaned {
		add(p.PR.Title)
	}

	return strings.Join(lines, "\n")
}
