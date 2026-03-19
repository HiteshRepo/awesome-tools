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

// BulletItem is one line in the status update with its source ticket IDs.
type BulletItem struct {
	Bullet  string   `json:"bullet"`
	Tickets []string `json:"tickets"`
}

// generateSummary calls Claude Haiku and returns structured bullet items.
// Falls back to raw titles if ANTHROPIC_API_KEY is unset or the call fails.
func generateSummary(ctx context.Context, view ConnectedView) ([]BulletItem, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "warning: ANTHROPIC_API_KEY not set, using raw titles")
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
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call anthropic: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("anthropic API %d: %s", resp.StatusCode, respBody)
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	if len(result.Content) == 0 {
		return nil, fmt.Errorf("empty response from anthropic: %s", respBody)
	}

	text := stripCodeFence(result.Content[0].Text)
	var items []BulletItem
	if err := json.Unmarshal([]byte(text), &items); err != nil {
		return nil, fmt.Errorf("parse JSON bullets: %w\nresponse was: %s", err, text)
	}
	return items, nil
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
4. Use plain English. No repo names.
5. In the "tickets" field, list ALL Jira IDs that contributed to that bullet.
   For bullets from GitHub PRs with no ticket, use an empty array.

EXAMPLES OF MERGING:
Input:
- [DP-1379] Remove progress-tracking from the Atlas flow
- [DP-1381] item-version job: remove updating progress-tracking records
- [DP-1382] delete progress-tracking job (app + iac)
- [DP-1380] discovery-svc: update recovery-points API to stop using progress-tracking data
Output JSON:
[{"bullet": "Removed progress-tracking system across Atlas flow, jobs, and APIs", "tickets": ["DP-1379", "DP-1381", "DP-1382", "DP-1380"]}]

INPUT FORMAT:
Each item is prefixed with its Jira ID in brackets (or no prefix for GitHub-only PRs).

OUTPUT FORMAT:
Return a JSON array only — no markdown, no explanation:
[{"bullet": "...", "tickets": ["ID-1", "ID-2"]}, ...]

ITEMS TO SUMMARISE:
`)

	if len(view.Linked) > 0 || len(view.NoGH) > 0 {
		sb.WriteString("Jira tickets:\n")
		for _, lt := range view.Linked {
			fmt.Fprintf(&sb, "- [%s] %s (%s)\n", lt.Ticket.ID, lt.Ticket.Summary, lt.Ticket.Status)
		}
		for _, t := range view.NoGH {
			fmt.Fprintf(&sb, "- [%s] %s (%s)\n", t.ID, t.Summary, t.Status)
		}
		sb.WriteString("\n")
	}

	if len(view.Orphaned) > 0 {
		sb.WriteString("GitHub PRs (no Jira ticket):\n")
		for _, p := range view.Orphaned {
			fmt.Fprintf(&sb, "- %s\n", p.PR.Title)
		}
	}

	return sb.String()
}

// fallbackSummary returns one BulletItem per raw title when the API is unavailable.
func fallbackSummary(view ConnectedView) []BulletItem {
	seen := make(map[string]bool)
	var items []BulletItem

	for _, lt := range view.Linked {
		if lt.Ticket.Summary != "" && !seen[lt.Ticket.Summary] {
			seen[lt.Ticket.Summary] = true
			items = append(items, BulletItem{Bullet: lt.Ticket.Summary, Tickets: []string{lt.Ticket.ID}})
		}
	}
	for _, t := range view.NoGH {
		if t.Summary != "" && !seen[t.Summary] {
			seen[t.Summary] = true
			items = append(items, BulletItem{Bullet: t.Summary, Tickets: []string{t.ID}})
		}
	}
	for _, p := range view.Orphaned {
		if p.PR.Title != "" && !seen[p.PR.Title] {
			seen[p.PR.Title] = true
			items = append(items, BulletItem{Bullet: p.PR.Title})
		}
	}
	return items
}

// stripCodeFence removes ```json ... ``` wrapping that models sometimes add.
func stripCodeFence(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		if idx := strings.Index(s, "\n"); idx != -1 {
			s = s[idx+1:]
		}
		if idx := strings.LastIndex(s, "```"); idx != -1 {
			s = s[:idx]
		}
	}
	return strings.TrimSpace(s)
}
