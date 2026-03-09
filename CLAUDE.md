# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make build          # Build all packages
make test           # Run all tests
make fmt            # Format code
make vet            # Static analysis
make lint           # Lint (requires golint installed)
make deps           # Download and tidy dependencies

# Run a single package's tests
go test -v ./pdf-reader/...
go test -v ./dttm/...
go test -v ./rate-limiter/...
go test -v ./go-struct-utils/...

# Coverage reports
make pdfreader-test-cov
make dttm-test-cov
make gostructutils-test-cov
```

## Architecture

This is a Go module (`github.com/HiteshRepo/awesome-tools`) — a collection of independent utility packages. Packages generally do not depend on each other, with the notable exception that `scraper` imports `rate-limiter`.

### Packages

**`rate-limiter/`** — Token bucket rate limiter using goroutines and channels. `NewRateLimiter(ctx, timeSpan, intervals)` creates a limiter where `intervals` actions are allowed per `timeSpan`. Callers use `Wait(ctx)` or `LimitCh(ctx)`.

**`scraper/`** — Concurrent web scraper with worker pool pattern. Wraps `rate-limiter` to control request rate. Inject custom scraping logic via `SetScrapeFunc(fn ScrapeFunc)`. The default implementation just returns HTTP status codes.

**`dttm/`** — Date/time parsing that auto-detects format via regex. `ParseTime(s)` handles all supported formats and always returns UTC. Multiple named `TimeFormat` constants are defined for formatting output.

**`go-struct-utils/`** — Three distinct struct→map conversion strategies with different tradeoffs:
- `StructToMapJSON`: marshal/unmarshal via JSON (numbers become float64)
- `StructToMapUsingReflection`: basic reflection, preserves Go types
- `StructToMapUsingAdvancedReflection`: full reflection with proper JSON tag handling

**`pdf-reader/`** — PDF text extraction. Supports page ranges, URL fetching, batch processing, and text search. Wraps `github.com/ledongthuc/pdf`.

**`atlassian/`** — MCP (Model Context Protocol) client for Jira, Confluence, and Rovo. Uses stdio MCP servers for Jira/Confluence and an HTTP MCP endpoint for Rovo. Confluence and Rovo startup failures are non-fatal (logged as warnings). Configuration is passed via `Config` struct with `ServerConfig` entries.
