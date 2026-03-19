package main

import (
	"fmt"
	"strings"
)

func formatMarkdown(from, to string, view ConnectedView) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "## Status Update: %s → %s\n\n", from, to)

	// ── Linked: Jira tickets with matching PRs ──────────────────────────────
	sb.WriteString("### Work Done\n")
	if len(view.Linked) == 0 && len(view.NoGH) == 0 {
		sb.WriteString("_No Jira tickets found._\n")
	}

	for _, lt := range view.Linked {
		t := lt.Ticket
		fmt.Fprintf(&sb, "\n#### [%s] %s (`%s`)\n", t.ID, t.Summary, t.Status)
		for _, p := range lt.PRs {
			state := prState(p.PR)
			fmt.Fprintf(&sb, "- **PR #%d** [%s](%s) — %s (`%s`)\n",
				p.PR.Number, p.PR.Title, p.PR.URL, state, p.Repo)
		}
	}

	// ── Jira tickets with no matching PR ────────────────────────────────────
	if len(view.NoGH) > 0 {
		sb.WriteString("\n### Jira Tickets Without Linked PRs\n")
		for _, t := range view.NoGH {
			fmt.Fprintf(&sb, "- **[%s]** %s (`%s`)\n", t.ID, t.Summary, t.Status)
		}
	}

	// ── PRs with no Jira reference ──────────────────────────────────────────
	if len(view.Orphaned) > 0 {
		sb.WriteString("\n### Untracked GitHub Work\n")
		sb.WriteString("> These PRs have no Jira ticket reference — consider creating tickets.\n\n")
		for _, p := range view.Orphaned {
			state := prState(p.PR)
			fmt.Fprintf(&sb, "- **PR #%d** [%s](%s) — %s (`%s`)\n",
				p.PR.Number, p.PR.Title, p.PR.URL, state, p.Repo)
		}
	}

	// ── Commits per repo ─────────────────────────────────────────────────────
	var reposWithCommits []RepoActivity
	for _, act := range view.AllActivity {
		if len(act.Commits) > 0 {
			reposWithCommits = append(reposWithCommits, act)
		}
	}
	if len(reposWithCommits) > 0 {
		sb.WriteString("\n### Commits\n")
		for _, act := range reposWithCommits {
			firstLine := strings.SplitN(act.Commits[0].Message, "\n", 2)[0]
			fmt.Fprintf(&sb, "- **%s** — %d commit(s) (latest: %q)\n",
				act.Repo, len(act.Commits), firstLine)
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
