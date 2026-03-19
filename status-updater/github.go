package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// repos returns the list of GitHub repos to query, read from GITHUB_REPOS
// (comma-separated, e.g. "owner/repo1,owner/repo2").
func repos() []string {
	val := os.Getenv("GITHUB_REPOS")
	if val == "" {
		return nil
	}
	var out []string
	for _, r := range strings.Split(val, ",") {
		if r := strings.TrimSpace(r); r != "" {
			out = append(out, r)
		}
	}
	return out
}

// PRSummary holds key fields from a GitHub pull request.
type PRSummary struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	State     string `json:"state"`
	Body      string `json:"body"`
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
	repoList := repos()
	if len(repoList) == 0 {
		fmt.Fprintln(os.Stderr, "warning: GITHUB_REPOS not set, no repos to scan")
	}

	for _, repo := range repoList {
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

// fetchPRs uses gh pr list (works for private org repos) and filters
// client-side by createdAt within [from, to].
func fetchPRs(ctx context.Context, repo, from, to string) ([]PRSummary, error) {
	args := []string{
		"pr", "list",
		"--repo", repo,
		"--author", "@me",
		"--state", "all",
		"--json", "number,title,url,state,body,createdAt,mergedAt",
		"--limit", "200",
	}

	out, err := exec.CommandContext(ctx, "gh", args...).Output()
	if err != nil {
		return nil, fmt.Errorf("gh pr list: %w", err)
	}

	var all []PRSummary
	if err := json.Unmarshal(out, &all); err != nil {
		return nil, fmt.Errorf("parse prs: %w", err)
	}

	// Filter client-side: keep PRs whose creation date falls within [from, to].
	var prs []PRSummary
	for _, pr := range all {
		if len(pr.CreatedAt) >= 10 {
			date := pr.CreatedAt[:10]
			if date >= from && date <= to {
				prs = append(prs, pr)
			}
		}
	}
	return prs, nil
}

// fetchCommits uses the REST API (since/until params) for reliable date filtering.
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
