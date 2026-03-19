package main

import (
	"fmt"
	"strings"

	"github.com/HiteshRepo/awesome-tools/atlassian"
)

func formatMarkdown(from, to string, tickets []atlassian.TicketSummary, activity []RepoActivity) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "## Status Update: %s → %s\n\n", from, to)

	sb.WriteString("### Jira Tickets\n")
	if len(tickets) == 0 {
		sb.WriteString("_No tickets found._\n")
	} else {
		for _, t := range tickets {
			fmt.Fprintf(&sb, "- **[%s]** %s (`%s`)\n", t.ID, t.Summary, t.Status)
		}
	}

	sb.WriteString("\n### GitHub Activity\n")

	hasActivity := false
	for _, act := range activity {
		if len(act.PRs) > 0 || len(act.Commits) > 0 {
			hasActivity = true
			break
		}
	}
	if !hasActivity {
		sb.WriteString("_No GitHub activity found._\n")
	}

	for _, act := range activity {
		if len(act.PRs) == 0 && len(act.Commits) == 0 {
			continue
		}

		fmt.Fprintf(&sb, "\n#### %s\n", act.Repo)

		for _, pr := range act.PRs {
			state := strings.ToLower(pr.State)
			if pr.MergedAt != "" {
				state = "merged"
			}
			fmt.Fprintf(&sb, "- **PR #%d** [%s](%s) — %s\n", pr.Number, pr.Title, pr.URL, state)
		}

		if len(act.Commits) > 0 {
			firstLine := strings.SplitN(act.Commits[0].Message, "\n", 2)[0]
			fmt.Fprintf(&sb, "- %d commit(s) (latest: %q)\n", len(act.Commits), firstLine)
		}
	}

	return sb.String()
}
