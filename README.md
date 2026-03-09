# Awesome Tools

A collection of useful Go utilities, SDKs, and development notes.

## Overview

This repository contains a set of Go tools and libraries designed to simplify common development workflows. Currently includes:

- **PDF Reader SDK** - A powerful Go library for PDF text extraction and processing
- **DTTM** - Date/time utilities for flexible parsing and formatting with multiple format support
- **Go Struct Utils** - Utilities for converting Go structs to maps using different approaches
- **Rate Limiter** - Token bucket rate limiter using goroutines and channels
- **Scraper** - Concurrent, rate-limited web scraper with a worker pool pattern
- **Atlassian** - MCP (Model Context Protocol) client for Jira, Confluence, and Rovo
- **CLI Commands** - Handy shell snippets for Go development
- **Pulumi** - Deployment notes and issue resolutions

## Project Structure

```
awesome-tools/
├── pdf-reader/          # PDF processing SDK
│   ├── reader.go
│   ├── utils.go
│   ├── examples/
│   └── README.md
├── dttm/                # Date/time utilities
│   ├── dttm.go
│   └── README.md
├── go-struct-utils/     # Struct to map conversion utilities
│   ├── utils.go
│   ├── utils_test.go
│   └── README.md
├── rate-limiter/        # Token bucket rate limiter
│   ├── rate_limiter.go
│   └── README.md
├── scraper/             # Concurrent rate-limited web scraper
│   ├── scraper.go
│   └── README.md
├── atlassian/           # MCP client for Jira, Confluence, Rovo
│   ├── client.go
│   ├── config.go
│   ├── jira.go
│   ├── confluence.go
│   ├── rovo.go
│   └── resources.go
├── cli-commands/        # Useful shell commands for Go development
│   └── commands-list.sh
├── pulumi/              # Pulumi deployment notes
│   └── issue#1.md
├── go.mod
├── Makefile
└── README.md
```

## Quick Start

### Prerequisites

- Go 1.24.1 or later

### Installation

Clone the repository:

```bash
git clone <repository-url>
cd awesome-tools
```

Install dependencies:

```bash
make deps
```

### Building

```bash
make build
```

### Testing

```bash
make test
```

## Available Tools

### PDF Reader SDK

A comprehensive Go SDK for reading and processing PDF documents.

**Key Features:**
- Extract text from PDF files
- Search text within PDFs
- Get document metadata
- Process PDFs from URLs
- Batch processing support

**Quick Example:**

```go
package main

import (
    "fmt"
    "log"

    pdfreader "github.com/HiteshRepo/awesome-tools/pdf-reader"
)

func main() {
    text, err := pdfreader.ExtractTextFromFile("document.pdf")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Extracted text:", text)
}
```

For detailed documentation, see [pdf-reader/README.md](pdf-reader/README.md).

---

### DTTM - Date/Time Utilities

A flexible Go package for parsing and formatting time strings in various formats.

**Key Features:**
- Multiple time format support (RFC3339, human-readable, custom formats)
- Automatic format detection during parsing
- UTC conversion for all parsed times

**Quick Example:**

```go
package main

import (
    "fmt"
    "log"

    "github.com/HiteshRepo/awesome-tools/dttm"
)

func main() {
    t, err := dttm.ParseTime("25-Dec-2023_15:30:45")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Standard: %s\n", dttm.FormatTo(t, dttm.Standard))
    fmt.Printf("Date Only: %s\n", dttm.FormatTo(t, dttm.DateOnly))
}
```

For detailed documentation, see [dttm/README.md](dttm/README.md).

---

### Go Struct Utils

Utilities for converting Go structs to maps using different approaches and strategies.

**Key Features:**
- Three conversion methods: JSON-based, basic reflection, advanced reflection
- Full JSON tag support (including `-` exclusion)
- Handles nested structs and pointers

**Quick Example:**

```go
package main

import (
    "fmt"
    "log"

    gostructutils "github.com/HiteshRepo/awesome-tools/go-struct-utils"
)

type Person struct {
    Name     string `json:"name"`
    Age      int    `json:"age"`
    Password string `json:"-"`
}

func main() {
    person := Person{Name: "John", Age: 30, Password: "secret"}
    result, err := gostructutils.StructToMapJSON(person)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Result: %+v\n", result)
    // Output: map[age:30 name:John]
}
```

For detailed documentation, see [go-struct-utils/README.md](go-struct-utils/README.md).

---

### Rate Limiter

A lightweight token bucket rate limiter using goroutines and channels.

**Key Features:**
- Limit N actions per time span
- Context-aware cancellation
- `Wait(ctx)` (blocking, recommended) and `LimitCh(ctx)` (channel-based) APIs

**Quick Example:**

```go
package main

import (
    "context"
    "fmt"
    "time"

    rlm "github.com/HiteshRepo/awesome-tools/rate-limiter"
)

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Allow 5 operations per second
    rl := rlm.NewRateLimiter(ctx, time.Second, 5)
    defer rl.Stop()

    for i := 0; i < 10; i++ {
        if rl.Wait(ctx) {
            fmt.Println("Do rate-limited work")
        }
    }
}
```

For detailed documentation, see [rate-limiter/README.md](rate-limiter/README.md).

---

### Scraper

Concurrent, rate-limited web scraper with a worker pool and injectable scrape logic.

**Key Features:**
- Worker pool for concurrency
- Configurable rate limiting (wraps `rate-limiter`)
- Custom scrape function via `SetScrapeFunc`
- Graceful cancellation via context

**Quick Example:**

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/HiteshRepo/awesome-tools/scraper"
)

func main() {
    ctx := context.Background()

    s := scraper.NewScraper(5, 3) // 5 rps, 3 workers
    s.SetRateLimit(ctx, time.Second, 5)

    urls := []string{"https://example.com", "https://golang.org"}
    resultsCh, err := s.ScrapeURLs(ctx, urls)
    if err != nil {
        panic(err)
    }

    for res := range resultsCh {
        if res.Error != nil {
            fmt.Printf("Failed: %s — %v\n", res.URL, res.Error)
        } else {
            fmt.Printf("Scraped: %s — %s\n", res.URL, res.Content)
        }
    }
}
```

For detailed documentation, see [scraper/README.md](scraper/README.md).

---

### Atlassian MCP Client

A Go client for interacting with Jira, Confluence, and Rovo via the Model Context Protocol (MCP).

**Key Features:**
- Jira and Confluence via stdio MCP servers
- Rovo via HTTP MCP endpoint
- Non-fatal startup for Confluence/Rovo (logged as warnings)
- Configuration via `Config` struct

**Quick Example:**

```go
package main

import (
    "github.com/HiteshRepo/awesome-tools/atlassian"
)

func main() {
    cfg := atlassian.Config{
        Jira: atlassian.ServerConfig{
            Command: "npx",
            Args:    []string{"-y", "@atlassian/mcp-jira"},
            Env:     map[string]string{"JIRA_TOKEN": "..."},
        },
        Rovo: atlassian.RovoConfig{
            URL:      "https://api.atlassian.com/mcp",
            Email:    "user@example.com",
            APIToken: "...",
            CloudID:  "<site-uuid>",
        },
    }

    client, err := atlassian.NewClient(cfg)
    if err != nil {
        panic(err)
    }
    defer client.Close()
}
```

---

### CLI Commands

Handy shell snippets for Go development tasks.

**Find unique versions of a package across all `go.mod` files in a repo:**

```bash
grep -R <package-name> --include="go.mod" . | awk '{print $2, $3}' | sort | uniq -c
```

See [cli-commands/commands-list.sh](cli-commands/commands-list.sh) for the full list.

---

### Pulumi Notes

Deployment notes and issue resolutions for Pulumi-based infrastructure.

- **[issue#1 — Helm Release Lock](pulumi/issue%231.md)**: How to diagnose and resolve `another operation (install/upgrade/rollback) is in progress` errors during Pulumi deployments.

---

## Development

### Available Make Commands

- `make build` - Build all packages
- `make test` - Run all tests
- `make fmt` - Format code
- `make vet` - Run go vet
- `make lint` - Run golint (requires golint installation)
- `make deps` - Download and tidy dependencies

### Per-Package Tests

```bash
go test -v ./pdf-reader/...
go test -v ./dttm/...
go test -v ./rate-limiter/...
go test -v ./go-struct-utils/...
```

### Coverage Reports

```bash
make pdfreader-test-cov
make dttm-test-cov
make gostructutils-test-cov
```

### Code Quality

```bash
make fmt && make vet && make lint && make test
```

## Dependencies

- `github.com/ledongthuc/pdf` - PDF parsing and text extraction
- `github.com/pkg/errors` - Enhanced error handling
- `github.com/mark3labs/mcp-go` - MCP client/server framework
- `github.com/docker/go-units` - Unit conversion utilities
- `github.com/spf13/cast` - Type casting utilities

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests and quality checks
5. Submit a pull request

## Support

For issues and questions, please open an issue in the repository.
