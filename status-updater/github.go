package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
)

// defaultRepos is the hardcoded list of GitHub repos to query.
// Edit this list to add or remove repos.
var defaultRepos = []string{
	"HiteshRepo/awesome-tools",
	// add more repos here
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
	var activities []RepoActivity
	var lastErr error

	for _, repo := range defaultRepos {
		act := RepoActivity{Repo: repo}

		prs, err := fetchPRs(ctx, repo, from, to)
		if err != nil {
			fmt.Printf("  warning: prs for %s: %v\n", repo, err)
			lastErr = err
		} else {
			act.PRs = prs
		}

		commits, err := fetchCommits(ctx, repo, from, to)
		if err != nil {
			fmt.Printf("  warning: commits for %s: %v\n", repo, err)
			lastErr = err
		} else {
			act.Commits = commits
		}

		activities = append(activities, act)
	}

	if len(activities) == 0 && lastErr != nil {
		return nil, lastErr
	}
	return activities, nil
}

func fetchPRs(ctx context.Context, repo, from, to string) ([]PRSummary, error) {
	args := []string{
		"pr", "list",
		"--repo", repo,
		"--author", "@me",
		"--state", "all",
		"--search", fmt.Sprintf("created:%s..%s", from, to),
		"--json", "number,title,url,state,createdAt,mergedAt",
		"--limit", "100",
	}

	out, err := exec.CommandContext(ctx, "gh", args...).Output()
	if err != nil {
		return nil, fmt.Errorf("gh pr list: %w", err)
	}

	var prs []PRSummary
	if err := json.Unmarshal(out, &prs); err != nil {
		return nil, fmt.Errorf("parse pr list: %w", err)
	}
	return prs, nil
}

func fetchCommits(ctx context.Context, repo, from, to string) ([]CommitSummary, error) {
	args := []string{
		"search", "commits",
		"--author", "@me",
		"--repo", repo,
		"--author-date", ">=" + from,
		"--author-date", "<=" + to,
		"--json", "sha,commit,url",
		"--limit", "100",
	}

	out, err := exec.CommandContext(ctx, "gh", args...).Output()
	if err != nil {
		return nil, fmt.Errorf("gh search commits: %w", err)
	}

	var raw []struct {
		SHA    string `json:"sha"`
		Commit struct {
			Message string `json:"message"`
		} `json:"commit"`
		URL string `json:"url"`
	}
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("parse commits: %w", err)
	}

	commits := make([]CommitSummary, len(raw))
	for i, r := range raw {
		commits[i] = CommitSummary{
			SHA:     r.SHA,
			Message: r.Commit.Message,
			URL:     r.URL,
		}
	}
	return commits, nil
}
