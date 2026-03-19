package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	from := flag.String("from", "", "Start date YYYY-MM-DD (required)")
	to := flag.String("to", "", "End date YYYY-MM-DD (required)")
	output := flag.String("output", "", "Write output to file (default: stdout)")
	flag.Parse()

	if *from == "" || *to == "" {
		fmt.Fprintln(os.Stderr, "Usage: status-updater --from YYYY-MM-DD --to YYYY-MM-DD [--output file]")
		os.Exit(1)
	}

	if _, err := time.Parse("2006-01-02", *from); err != nil {
		log.Fatalf("invalid --from date: %v", err)
	}
	if _, err := time.Parse("2006-01-02", *to); err != nil {
		log.Fatalf("invalid --to date: %v", err)
	}

	ctx := context.Background()

	tickets, err := fetchJiraTickets(ctx, *from, *to)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: jira fetch failed: %v\n", err)
	}

	activity, err := fetchGitHubActivity(ctx, *from, *to)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: github fetch failed: %v\n", err)
	}

	md := formatMarkdown(*from, *to, tickets, activity)

	if *output != "" {
		if err := os.WriteFile(*output, []byte(md), 0644); err != nil {
			log.Fatalf("write output: %v", err)
		}
	} else {
		fmt.Print(md)
	}
}
