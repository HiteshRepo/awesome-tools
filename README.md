# Awesome Tools

A collection of Go utilities, Python libraries, and developer tooling.

## Packages

### Go

| Package | Description |
|---------|-------------|
| `rate-limiter` | Token bucket rate limiter using goroutines and channels |
| `scraper` | Concurrent, rate-limited web scraper with worker pool and injectable scrape logic |
| `dttm` | Date/time parsing with auto-detected format and UTC output |
| `go-struct-utils` | Three struct→map strategies: JSON, basic reflection, advanced reflection |
| `pdf-reader` | PDF text extraction with page ranges, URL fetching, and search |
| `atlassian` | MCP client for Jira, Confluence, and Rovo |
| `status-updater` | CLI that generates a markdown status report from Jira + GitHub activity, optionally summarized via Claude Haiku |

### Python (`python/`)

| Package | Description |
|---------|-------------|
| `std-llm-client` | Abstract LLM client with provider implementations |
| `std-embeddings` | Abstract embedding provider (OpenAI, Voyage, local) |
| `std-vector-store` | Abstract vector store interface with pluggable backends |
| `std-mcp-utils` | MCP server utilities — tool registration, types, request handling |
| `std-rag` | RAG pipeline wiring together LLM, embeddings, vector store, and chunker |

### Notes

| Directory | Description |
|-----------|-------------|
| `cli-commands/` | Shell snippets for Go development tasks |
| `pulumi/` | Deployment notes and issue resolutions |

---

## Project Structure

```
awesome-tools/
├── rate-limiter/
├── scraper/
├── dttm/
├── go-struct-utils/
├── pdf-reader/
├── atlassian/
├── status-updater/
├── python/
│   ├── std-llm-client/
│   ├── std-embeddings/
│   ├── std-vector-store/
│   ├── std-mcp-utils/
│   └── std-rag/
├── cli-commands/
├── pulumi/
├── go.mod
└── Makefile
```

---

## Quick Start

**Prerequisites:** Go 1.24.1+, Python 3.11+ (for Python packages)

```bash
# Install Go dependencies
make deps

# Build all Go packages
make build

# Run all Go tests
make test
```

---

## Go Package Details

### rate-limiter

Allow N operations per time span. `Wait(ctx)` blocks; `LimitCh(ctx)` returns a channel.

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

rl := rlm.NewRateLimiter(ctx, time.Second, 5) // 5 ops/sec
defer rl.Stop()

for i := 0; i < 10; i++ {
    if rl.Wait(ctx) {
        fmt.Println("rate-limited work")
    }
}
```

---

### scraper

Worker pool scraper that wraps `rate-limiter`. Inject custom logic via `SetScrapeFunc`.

```go
s := scraper.NewScraper(5, 3) // 5 rps, 3 workers
s.SetRateLimit(ctx, time.Second, 5)

resultsCh, _ := s.ScrapeURLs(ctx, []string{"https://example.com"})
for res := range resultsCh {
    fmt.Printf("%s — %s\n", res.URL, res.Content)
}
```

---

### dttm

Auto-detects format, always returns UTC.

```go
t, _ := dttm.ParseTime("25-Dec-2023_15:30:45")
fmt.Println(dttm.FormatTo(t, dttm.Standard))
```

---

### go-struct-utils

Three conversion strategies with different type-preservation tradeoffs.

```go
type Person struct {
    Name     string `json:"name"`
    Password string `json:"-"`
}

result, _ := gostructutils.StructToMapJSON(Person{Name: "John", Password: "secret"})
// map[name:John]  — Password excluded by json:"-"
```

---

### pdf-reader

Extract text, search content, fetch from URL, batch process.

```go
text, _ := pdfreader.ExtractTextFromFile("document.pdf")
fmt.Println(text)
```

See [pdf-reader/README.md](pdf-reader/README.md).

---

### atlassian

MCP client for Jira (stdio), Confluence (stdio), and Rovo (HTTP). Confluence and Rovo startup failures are non-fatal.

```go
cfg := atlassian.Config{
    Jira: atlassian.ServerConfig{
        Command: "npx",
        Args:    []string{"-y", "@atlassian/mcp-jira"},
        Env:     map[string]string{"JIRA_TOKEN": "..."},
    },
}
client, _ := atlassian.NewClient(cfg)
defer client.Close()
```

---

### status-updater

Generates a markdown status report for a date range by pulling Jira tickets and GitHub PRs, then optionally rewriting them into clean past-tense bullets via Claude Haiku.

**Setup:**

```bash
cp status-updater/sample.env .env
# fill in ANTHROPIC_API_KEY, GITHUB_REPOS, JIRA_* values
source .env
```

**Run:**

```bash
go run ./status-updater --from 2024-01-01 --to 2024-01-07
go run ./status-updater --from 2024-01-01 --to 2024-01-07 --output report.md
```

Jira source priority: `acli` (if on PATH) > REST API (`JIRA_URL` + `JIRA_EMAIL` + `JIRA_API_TOKEN`) > MCP server.

---

## Python Package Details

All packages live under `python/` and follow a provider-abstraction pattern.

### std-llm-client

Abstract `LLMClient` base with concrete provider implementations.

```python
from std_llm_client import LLMClient, Message
```

### std-embeddings

Abstract `EmbeddingProvider` supporting OpenAI, Voyage, and local models.

```python
from std_embeddings import EmbeddingProvider, EmbeddingConfig
```

### std-vector-store

Abstract `VectorStore` with pluggable backends. Operates on `Document` and `SearchResult` types.

```python
from std_vector_store import VectorStore, Document
```

### std-mcp-utils

Utilities for building MCP servers: tool registration, type definitions, request handling.

```python
from std_mcp_utils import MCPServer, tool
```

### std-rag

End-to-end RAG pipeline composing `std-llm-client`, `std-embeddings`, `std-vector-store`, and a text chunker.

```python
pipeline = RAGPipeline(llm=client, embedder=embedder, store=store)
response = pipeline.query("What is the refund policy?")
```

---

## Development

### Make Commands

```bash
make build              # Build all Go packages
make test               # Run all Go tests
make fmt                # Format Go code
make vet                # Run go vet
make lint               # Run golint
make deps               # Download and tidy dependencies

# Per-package tests
go test -v ./pdf-reader/...
go test -v ./dttm/...
go test -v ./rate-limiter/...
go test -v ./go-struct-utils/...

# Coverage reports
make pdfreader-test-cov
make dttm-test-cov
make gostructutils-test-cov
```

### Dependencies

- `github.com/ledongthuc/pdf` — PDF parsing
- `github.com/pkg/errors` — enhanced error handling
- `github.com/mark3labs/mcp-go` — MCP client/server framework
- `github.com/docker/go-units` — unit conversion
- `github.com/spf13/cast` — type casting
