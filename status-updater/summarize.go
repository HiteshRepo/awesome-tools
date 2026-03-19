package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("call anthropic: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("anthropic API %d: %s", resp.StatusCode, respBody)
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}
	if len(result.Content) == 0 {
		return "", fmt.Errorf("empty response from anthropic: %s", respBody)
	}

	return strings.TrimSpace(result.Content[0].Text), nil
}

func buildPrompt(view ConnectedView) string {
	var sb strings.Builder

	sb.WriteString(`You are writing a developer's weekly status update.
Convert the raw items below into a SHORT list of clean past-tense bullet points.

RULES:
1. MERGE related items into ONE bullet.
   - Multiple dashboard updates → one bullet about dashboard work.
   - Multiple items removing the same system → one bullet summarising the removal.
   - Multiple items creating the same credential/principal → one bullet.
2. SKIP trivial items (e.g. "fix lint failure", "fix typo", "bump version").
3. ADD brief context in parentheses when a title alone is unclear.
   Example: "Support query-service key for discovery routing"
         → "Added geo-based routing key support for discovery service (replaces regional routing)"
4. Use plain English. No ticket IDs, PR numbers, or repo names.
5. Output ONLY the bullet lines, each starting with "- ".

EXAMPLES OF MERGING:
Input:
- update metadata cp dashboard
- update dashboard mar13
- update routing svc widgets - mar9
Output:
- Updated monitoring dashboards for item catalog and routing services

Input:
- Remove progress-tracking from the Atlas flow
- delete progress-tracking job (app + iac)
- item-version job: remove updating progress-tracking records
- discovery-svc: update recovery-points API to stop using progress-tracking data
Output:
- Removed progress-tracking system across Atlas flow, jobs, and APIs

Input:
- fix: use stack.getOutput() for dev-only azure-oidc optional outputs
- create Azure app registration for integ test service principal
- Separate service principal for CD integ tests
Output:
- Set up dedicated Azure service principal and app registration for integration test CI

Input:
- ATLAS: Minimal devex for the data pipeline
- ATLAS: DataBricks query feasibility spike
Output:
- Explored DataBricks query feasibility and developer experience improvements for the data pipeline

ITEMS TO SUMMARISE:
`)

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
