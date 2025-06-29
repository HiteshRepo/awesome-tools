package scraper

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	rlm "github.com/hiteshrepo/awesome-tools/rate-limiter"
)

type ScrapeFunc func(ctx context.Context, url string) Result

const (
	ScrapeTimeoutDuration = 10 * time.Second
)

type RateLimitedScraper struct {
	client     *http.Client
	rlm        *rlm.RateLimiter
	maxWorkers int
	scrapeFunc ScrapeFunc
}

func NewScraper(rps int, maxWorkers int) *RateLimitedScraper {
	s := &RateLimitedScraper{
		client:     &http.Client{Timeout: ScrapeTimeoutDuration},
		maxWorkers: maxWorkers,
	}

	// Use default scrape function
	s.scrapeFunc = s.scrapeURL
	return s
}

// Sets the rate limit of the scraper.
func (s *RateLimitedScraper) SetRateLimit(
	ctx context.Context,
	timeSpan time.Duration,
	intervals int,
) {
	r := rlm.NewRateLimiter(ctx, timeSpan, intervals)
	s.rlm = &r
}

// Allow callers to set a custom scrape function
func (s *RateLimitedScraper) SetScrapeFunc(fn ScrapeFunc) {
	s.scrapeFunc = fn
}

func (s *RateLimitedScraper) ScrapeURLs(
	ctx context.Context,
	urls []string) (<-chan Result, error) {
	if s.rlm == nil {
		return nil, fmt.Errorf("rate limiter not set")
	}

	results := make(chan Result, len(urls))
	urlChan := make(chan string, len(urls))

	go func() {
		defer close(urlChan)
		for _, url := range urls {
			select {
			case urlChan <- url:
			case <-ctx.Done():
				// s.rlm.Stop() is not required because the expectation is that the same ctx,
				// would have been passed to the rate limiter while the latter's initialization.
				return
			}
		}
	}()

	var wg sync.WaitGroup
	for i := 0; i < s.maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for url := range urlChan {
				select {
				case <-s.rlm.LimitCh(ctx):
					result := s.scrapeFunc(ctx, url)
					select {
					case results <- result:
					case <-ctx.Done():
						// s.rlm.Stop() is not required because the expectation is that the same ctx,
						// would have been passed to the rate limiter while the latter's initialization.
						return
					}
				case <-ctx.Done():
					// s.rlm.Stop() is not required because the expectation is that the same ctx,
					// would have been passed to the rate limiter while the latter's initialization.
					return
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
		s.rlm.Stop()
	}()

	return results, nil
}

type Result struct {
	URL     string
	Content string
	Error   error
}

func (s *RateLimitedScraper) scrapeURL(ctx context.Context, url string) Result {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return Result{URL: url, Error: err}
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return Result{URL: url, Error: err}
	}
	defer resp.Body.Close()

	// Read content (simplified)
	return Result{URL: url, Content: fmt.Sprintf("Status: %d", resp.StatusCode)}
}
