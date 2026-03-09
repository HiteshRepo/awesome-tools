package atlassian

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// ConfluencePage holds the content of a Confluence page or search result.
type ConfluencePage struct {
	Title    string
	URL      string
	Excerpt  string
	IsFolder bool
}

// FetchPage fetches a single Confluence page by its URL or numeric page ID.
func (c *Client) FetchPage(ctx context.Context, pageURL string) (*ConfluencePage, error) {
	if c.confluence == nil {
		return nil, fmt.Errorf("confluence mcp not available")
	}
	return c.fetchConfluencePage(ctx, pageURL)
}

// GetConfluenceContext fetches pages explicitly linked in the ticket.
func (c *Client) GetConfluenceContext(ctx context.Context, ticket *Ticket) ([]ConfluencePage, error) {
	if c.confluence == nil {
		return nil, fmt.Errorf("confluence mcp not available")
	}

	var pages []ConfluencePage
	for _, link := range ticket.ConfluenceLinks {
		page, err := c.fetchConfluencePage(ctx, link)
		if err == nil {
			pages = append(pages, *page)
		}
	}
	return pages, nil
}

// SearchConfluenceBySpace searches Confluence pages within a specific space.
// spaceKey is the Confluence space key (e.g. "DEV"). Use an empty string to
// search across all spaces.
func (c *Client) SearchConfluenceBySpace(ctx context.Context, spaceKey, query string, limit int) ([]ConfluencePage, error) {
	if c.confluence == nil {
		return nil, fmt.Errorf("confluence mcp not available")
	}

	cql := BuildCQL(spaceKey, query)
	req := mcp.CallToolRequest{}
	req.Params.Name = "confluence_search"
	req.Params.Arguments = map[string]any{
		"query": cql,
		"limit": limit,
	}

	resp, err := c.confluence.CallTool(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("confluence_search: %w", err)
	}

	return parseConfluenceSearchResults(extractText(resp)), nil
}

// BuildCQL constructs a Confluence CQL query string.
// Space keys are short uppercase identifiers (e.g. "DP"); anything with
// lowercase letters or spaces is treated as a space title.
func BuildCQL(space, query string) string {
	if space == "" {
		return fmt.Sprintf(`text ~ "%s"`, query)
	}
	spaceField := "space"
	if strings.ContainsAny(space, " ") || space != strings.ToUpper(space) {
		spaceField = "space.title"
	}
	return fmt.Sprintf(`%s = "%s" AND text ~ "%s"`, spaceField, space, query)
}

func (c *Client) fetchConfluencePage(ctx context.Context, pageIDOrURL string) (*ConfluencePage, error) {
	req := mcp.CallToolRequest{}
	req.Params.Name = "confluence_get_page"
	req.Params.Arguments = map[string]any{
		"page_id": extractPageID(pageIDOrURL),
	}

	resp, err := c.confluence.CallTool(ctx, req)
	if err != nil {
		return nil, err
	}

	return parseConfluencePage(extractText(resp)), nil
}

// extractPageID pulls the numeric page ID out of a Confluence URL.
// Returns the input unchanged if it's already a bare ID.
func extractPageID(pageIDOrURL string) string {
	if !strings.Contains(pageIDOrURL, "/") {
		return pageIDOrURL
	}
	parts := strings.Split(pageIDOrURL, "/")
	for i, p := range parts {
		if p == "pages" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return pageIDOrURL
}

func parseConfluencePage(raw string) *ConfluencePage {
	var payload struct {
		Metadata struct {
			Title   string `json:"title"`
			URL     string `json:"url"`
			Content struct {
				Value string `json:"value"`
			} `json:"content"`
		} `json:"metadata"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err == nil && payload.Metadata.Content.Value != "" {
		excerpt := payload.Metadata.Content.Value
		if len(excerpt) > 8000 {
			excerpt = excerpt[:8000]
		}
		return &ConfluencePage{
			Title:   payload.Metadata.Title,
			URL:     payload.Metadata.URL,
			Excerpt: excerpt,
		}
	}
	excerpt := raw
	if len(excerpt) > 8000 {
		excerpt = excerpt[:8000]
	}
	return &ConfluencePage{Excerpt: excerpt}
}

func parseConfluenceSearchResults(raw string) []ConfluencePage {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "[]" || raw == "null" {
		return nil
	}

	var items []struct {
		Title   string `json:"title"`
		Type    string `json:"type"`
		URL     string `json:"url"`
		Content struct {
			Value string `json:"value"`
		} `json:"content"`
	}
	if err := json.Unmarshal([]byte(raw), &items); err == nil {
		var pages []ConfluencePage
		for _, item := range items {
			excerpt := item.Content.Value
			if len(excerpt) > 2000 {
				excerpt = excerpt[:2000]
			}
			pages = append(pages, ConfluencePage{
				Title:    item.Title,
				URL:      item.URL,
				Excerpt:  excerpt,
				IsFolder: item.Type == "folder",
			})
		}
		return pages
	}

	var pages []ConfluencePage
	for _, block := range strings.Split(raw, "\n\n") {
		if block = strings.TrimSpace(block); block != "" {
			pages = append(pages, ConfluencePage{Excerpt: block})
		}
	}
	return pages
}
