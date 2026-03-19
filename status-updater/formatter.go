package main

import (
	"fmt"
	"strings"
)

func formatMarkdown(from, to string, summary string, view ConnectedView) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "## Status Update: %s → %s\n\n", from, to)
	sb.WriteString(summary)
	sb.WriteString("\n")

	// Untracked PRs — suggest creating tickets.
	if len(view.Orphaned) > 0 {
		sb.WriteString("\n---\n")
		sb.WriteString("**Untracked work** (no Jira ticket found — consider creating tickets):\n")
		for _, p := range view.Orphaned {
			state := prState(p.PR)
			fmt.Fprintf(&sb, "- [%s](%s) — %s (`%s`)\n", p.PR.Title, p.PR.URL, state, p.Repo)
		}
	}

	return sb.String()
}

func prState(pr PRSummary) string {
	if pr.MergedAt != "" {
		return "merged"
	}
	return strings.ToLower(pr.State)
}
