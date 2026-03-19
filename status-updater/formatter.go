package main

import (
	"fmt"
	"strings"
)

func formatMarkdown(from, to string, bullets []BulletItem, view ConnectedView) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "## Status Update: %s → %s\n\n", from, to)

	for _, item := range bullets {
		if len(item.Tickets) > 0 {
			fmt.Fprintf(&sb, "- %s (%s)\n", item.Bullet, strings.Join(item.Tickets, ", "))
		} else {
			fmt.Fprintf(&sb, "- %s\n", item.Bullet)
		}
	}

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
