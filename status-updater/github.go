package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// defaultRepos is the hardcoded list of GitHub repos to query.
// Edit this list to add or remove repos.
var defaultRepos = []string{
	"Veeam-VDC/vdc-shared-data-plane",
	"Veeam-VDC/vdc-shared-utils",
	"Veeam-VDC/control-plane-backend",
	"Veeam-VDC/control-plane-platform",
}

// PRSummary holds key fields from a GitHub pull request.
type PRSummary struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	State     string `json:"state"`
	CreatedAt string `json:"createdAt"`
	MergedAt  string `json:"mergedAt"`
}

// CommitSummary holds key fields from a GitHub commit.
type CommitSummary struct {
	SHA     string
	Message string
	URL     string
}

// RepoActivity groups PRs and commits for a single repo.
type RepoActivity struct {
	Repo    string
	PRs     []PRSummary
	Commits []CommitSummary
}

func fetchGitHubActivity(ctx context.Context, from, to string) ([]RepoActivity, error) {
	login, err := fetchGitHubLogin(ctx)
	if err != nil {
		return nil, fmt.Errorf("get github login: %w", err)
	}

	var activities []RepoActivity
	for _, repo := range defaultRepos {
		act := RepoActivity{Repo: repo}

		prs, err := fetchPRs(ctx, repo, from, to)
		if err != nil {
			fmt.Printf("  warning: prs for %s: %v\n", repo, err)
		} else {
			act.PRs = prs
		}

		commits, err := fetchCommits(ctx, repo, from, to, login)
		if err != nil {
			fmt.Printf("  warning: commits for %s: %v\n", repo, err)
		} else {
			act.Commits = commits
		}

		activities = append(activities, act)
	}

	return activities, nil
}

// fetchGitHubLogin returns the authenticated user's GitHub login.
func fetchGitHubLogin(ctx context.Context) (string, error) {
	out, err := exec.CommandContext(ctx, "gh", "api", "user", "--jq", ".login").Output()
	if err != nil {
		return "", fmt.Errorf("gh api user: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// fetchPRs uses gh search prs which has reliable --created date filtering.
func fetchPRs(ctx context.Context, repo, from, to string) ([]PRSummary, error) {
	args := []string{
		"search", "prs",
		"--author", "@me",
		"--repo", repo,
		"--created", fmt.Sprintf("%s..%s", from, to),
		"--json", "number,title,url,state,createdAt,mergedAt",
		"--limit", "100",
	}

	out, err := exec.CommandContext(ctx, "gh", args...).Output()
	if err != nil {
		return nil, fmt.Errorf("gh search prs: %w", err)
	}

	var prs []PRSummary
	if err := json.Unmarshal(out, &prs); err != nil {
		return nil, fmt.Errorf("parse prs: %w", err)
	}
	return prs, nil
}

// fetchCommits uses the REST API (since/until params) which is more reliable
// than gh search commits for date-bounded queries.
func fetchCommits(ctx context.Context, repo, from, to, login string) ([]CommitSummary, error) {
	endpoint := fmt.Sprintf(
		"repos/%s/commits?since=%sT00:00:00Z&until=%sT23:59:59Z&author=%s&per_page=100",
		repo, from, to, login,
	)
	out, err := exec.CommandContext(ctx, "gh", "api", endpoint, "--paginate").Output()
	if err != nil {
		return nil, fmt.Errorf("gh api commits: %w", err)
	}

	items, err := flattenJSONArrays(out)
	if err != nil {
		return nil, fmt.Errorf("parse commits: %w", err)
	}

	var commits []CommitSummary
	for _, item := range items {
		var c struct {
			SHA    string `json:"sha"`
			Commit struct {
				Message string `json:"message"`
			} `json:"commit"`
			HTMLURL string `json:"html_url"`
		}
		if err := json.Unmarshal(item, &c); err != nil {
			continue
		}
		commits = append(commits, CommitSummary{
			SHA:     c.SHA,
			Message: c.Commit.Message,
			URL:     c.HTMLURL,
		})
	}
	return commits, nil
}

// flattenJSONArrays merges paginated gh api output (multiple JSON arrays
// concatenated) into a single slice of raw messages.
func flattenJSONArrays(data []byte) ([]json.RawMessage, error) {
	var all []json.RawMessage
	dec := json.NewDecoder(bytes.NewReader(data))
	for {
		var arr []json.RawMessage
		if err := dec.Decode(&arr); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		all = append(all, arr...)
	}
	return all, nil
}
