# Rate-Limited Web Scraper

This Go package provides a concurrent, rate-limited web scraper using worker pools and a customizable rate limiter.

## Features
- Concurrency via worker pool
- Configurable rate limits
- Graceful cancellation via context
- Simple interface for scraping multiple URLs

## Installation
```bash
go get github.com/hiteshrepo/awesome-tools/scraper
```

Also depends on the rate-limiter package from the same repo.

## Usage

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/hiteshrepo/awesome-tools/scraper"
)

func main() {
	ctx := context.Background()

	s := scraper.NewScraper(rps=5, maxWorkers=3)
	s.SetRateLimit(ctx, 1*time.Second, 5) // 5 requests per second

	urls := []string{
		"https://example.com",
		"https://golang.org",
	}

	resultsCh, err := s.ScrapeURLs(ctx, urls)
	if err != nil {
		panic(err)
	}

	for res := range resultsCh {
		if res.Error != nil {
			fmt.Printf("Failed to scrape %s: %v\n", res.URL, res.Error)
		} else {
			fmt.Printf("Scraped %s: %s\n", res.URL, res.Content)
		}
	}
}

```

## API

### NewScraper(rps int, maxWorkers int)
Creates a new scraper instance with the given requests-per-second and number of concurrent workers.

### SetRateLimit(ctx, timeSpan, intervals)
Configures the rate limiter. For example: SetRateLimit(ctx, time.Second, 5) allows 5 requests per second.

### ScrapeURLs(ctx, urls []string) (<-chan Result, error)
Starts scraping the provided URLs concurrently, respecting the rate limit. Returns a channel of Result.

### Result Struct
```go
type Result struct {
	URL     string
	Content string
	Error   error
}
```

## Custom Scrape Function

```go
s.SetScrapeFunc(func(ctx context.Context, url string) scraper.Result {
	// Custom fetch + parse logic here
	// Example: return HTML title
	resp, err := http.Get(url)
	if err != nil {
		return scraper.Result{URL: url, Error: err}
	}
	defer resp.Body.Close()

	return scraper.Result{
		URL:     url,
		Content: fmt.Sprintf("Custom response status: %d", resp.StatusCode),
	}
})
```

If not provided, the default implementation (scrapeURL) is used, which simply returns the HTTP status code.

## Notes
- The actual scraping logic is minimal (GET with status code).
- Rate limiting is done using a token bucket-style limiter from the rate-limiter package.
- Customize scrapeURL() for more advanced parsing.