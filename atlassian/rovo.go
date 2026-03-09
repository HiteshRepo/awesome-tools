package atlassian

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// ListRovoTools returns all tool names exposed by the Atlassian official MCP.
func (c *Client) ListRovoTools(ctx context.Context) ([]string, error) {
	if c.rovo == nil {
		return nil, fmt.Errorf("rovo mcp not available")
	}
	result, err := c.rovo.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return nil, err
	}
	names := make([]string, len(result.Tools))
	for i, t := range result.Tools {
		names[i] = t.Name
	}
	return names, nil
}

// SearchRovo uses the Atlassian official MCP's Rovo semantic search tool.
func (c *Client) SearchRovo(ctx context.Context, query string) ([]ConfluencePage, error) {
	if c.rovo == nil {
		return nil, fmt.Errorf("rovo mcp not available")
	}

	req := mcp.CallToolRequest{}
	req.Params.Name = "search"
	req.Params.Arguments = map[string]any{
		"query":   query,
		"cloudId": c.cfg.Rovo.CloudID,
	}

	resp, err := c.rovo.CallTool(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("rovo search: %w", err)
	}

	raw := extractText(resp)
	fmt.Printf("  [rovo raw] %s\n", truncate(raw, 300))

	return parseRovoResults(raw), nil
}

func parseRovoResults(raw string) []ConfluencePage {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "[]" || raw == "null" {
		return nil
	}

	var items []struct {
		Title   string `json:"title"`
		URL     string `json:"url"`
		Excerpt string `json:"excerpt"`
		Content struct {
			Value string `json:"value"`
		} `json:"content"`
		Type string `json:"type"`
	}
	if err := json.Unmarshal([]byte(raw), &items); err == nil {
		var pages []ConfluencePage
		for _, item := range items {
			excerpt := item.Excerpt
			if excerpt == "" {
				excerpt = item.Content.Value
			}
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

	excerpt := raw
	if len(excerpt) > 2000 {
		excerpt = excerpt[:2000]
	}
	return []ConfluencePage{{Excerpt: excerpt}}
}
