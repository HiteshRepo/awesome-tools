package main

import (
	"regexp"

	"github.com/HiteshRepo/awesome-tools/atlassian"
)

// jiraKeyRE matches Jira-style ticket IDs like CP-123, DP-4567.
var jiraKeyRE = regexp.MustCompile(`\b([A-Z][A-Z0-9]+-\d+)\b`)

// PRWithRepo pairs a PR with the repo it came from.
type PRWithRepo struct {
	PR   PRSummary
	Repo string
}

// LinkedTicket is a Jira ticket together with all PRs that reference it.
type LinkedTicket struct {
	Ticket atlassian.TicketSummary
	PRs    []PRWithRepo
}

// ConnectedView is the result of linking Jira tickets to GitHub PRs.
type ConnectedView struct {
	// Linked contains Jira tickets that have at least one referencing PR.
	Linked []LinkedTicket
	// NoGH contains Jira tickets with no matching PR reference.
	NoGH []atlassian.TicketSummary
	// Orphaned contains PRs that reference no known Jira ticket.
	Orphaned []PRWithRepo
	// AllActivity is the raw per-repo data (for commit display).
	AllActivity []RepoActivity
}

// connect links Jira tickets to GitHub PRs by searching for Jira IDs
// in PR titles and bodies.
func connect(tickets []atlassian.TicketSummary, activity []RepoActivity) ConnectedView {
	// Index tickets by ID for O(1) lookup.
	index := make(map[string]*LinkedTicket, len(tickets))
	for i := range tickets {
		index[tickets[i].ID] = &LinkedTicket{Ticket: tickets[i]}
	}

	var orphaned []PRWithRepo

	for _, act := range activity {
		for _, pr := range act.PRs {
			refs := extractJiraKeys(pr.Title + " " + pr.Body)

			matched := false
			for _, key := range refs {
				if lt, ok := index[key]; ok {
					lt.PRs = append(lt.PRs, PRWithRepo{PR: pr, Repo: act.Repo})
					matched = true
				}
			}
			if !matched {
				orphaned = append(orphaned, PRWithRepo{PR: pr, Repo: act.Repo})
			}
		}
	}

	var linked []LinkedTicket
	var noGH []atlassian.TicketSummary
	for _, t := range tickets {
		lt := index[t.ID]
		if len(lt.PRs) > 0 {
			linked = append(linked, *lt)
		} else {
			noGH = append(noGH, t)
		}
	}

	return ConnectedView{
		Linked:      linked,
		NoGH:        noGH,
		Orphaned:    orphaned,
		AllActivity: activity,
	}
}

// extractJiraKeys returns deduplicated Jira ticket IDs found in text.
func extractJiraKeys(text string) []string {
	matches := jiraKeyRE.FindAllString(text, -1)
	seen := make(map[string]bool, len(matches))
	var keys []string
	for _, m := range matches {
		if !seen[m] {
			seen[m] = true
			keys = append(keys, m)
		}
	}
	return keys
}
