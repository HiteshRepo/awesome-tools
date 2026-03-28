package atlassian

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	mcpgo "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

// Client wraps MCP connections to Jira, Confluence, and Rovo.
type Client struct {
	jira          *mcpgo.Client
	confluence    *mcpgo.Client
	rovo          *mcpgo.Client
	claudeCodeMCP *mcpgo.Client // HTTP MCP server authenticated via Claude Code OAuth
	cfg           Config
}

// NewClient starts the configured MCP servers. Individual server failures
// are non-fatal and are printed as warnings.
func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	c := &Client{cfg: cfg}

	if cfg.ClaudeCodeMCP.ServerName != "" {
		if err := c.startClaudeCodeMCP(ctx); err != nil {
			fmt.Printf("  ⚠ claude code mcp unavailable: %v\n", err)
		}
	}

	if cfg.Jira.Command != "" {
		if err := c.startJira(ctx); err != nil {
			fmt.Printf("  ⚠ jira mcp unavailable: %v\n", err)
		}
	}

	if cfg.Confluence.Command != "" {
		if err := c.startConfluence(ctx); err != nil {
			fmt.Printf("  ⚠ confluence mcp unavailable: %v\n", err)
		}
	}

	if cfg.Rovo.URL != "" {
		if err := c.startRovo(ctx); err != nil {
			fmt.Printf("  ⚠ rovo mcp unavailable: %v\n", err)
		}
	}

	return c, nil
}

// Close shuts down all active MCP connections.
func (c *Client) Close() {
	if c.claudeCodeMCP != nil {
		c.claudeCodeMCP.Close()
	}
	if c.jira != nil {
		c.jira.Close()
	}
	if c.confluence != nil {
		c.confluence.Close()
	}
}

func (c *Client) startClaudeCodeMCP(ctx context.Context) error {
	cfg := c.cfg.ClaudeCodeMCP

	token, err := loadClaudeToken(cfg.CredentialsPath, cfg.ServerName)
	if err != nil {
		return fmt.Errorf("load claude code token: %w", err)
	}

	// Derive the server URL from the stored credential (always mcp suffix).
	// We need the URL — read it from credentials rather than hard-coding it.
	serverURL, err := claudeCodeMCPServerURL(cfg.CredentialsPath, cfg.ServerName)
	if err != nil {
		return fmt.Errorf("resolve server URL: %w", err)
	}

	t, err := transport.NewStreamableHTTP(serverURL,
		transport.WithHTTPHeaders(map[string]string{
			"Authorization": "Bearer " + token,
		}),
	)
	if err != nil {
		return fmt.Errorf("create transport: %w", err)
	}

	client := mcpgo.NewClient(t)
	if err := initMCP(ctx, client); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	c.claudeCodeMCP = client
	return nil
}

func (c *Client) startJira(ctx context.Context) error {
	cfg := c.cfg.Jira
	env := buildEnv(cfg.Env)

	client, err := mcpgo.NewStdioMCPClient(cfg.Command, env, cfg.Args...)
	if err != nil {
		return err
	}

	if err := initMCP(ctx, client); err != nil {
		return err
	}

	c.jira = client
	return nil
}

func (c *Client) startConfluence(ctx context.Context) error {
	cfg := c.cfg.Confluence
	env := buildEnv(cfg.Env)

	client, err := mcpgo.NewStdioMCPClient(cfg.Command, env, cfg.Args...)
	if err != nil {
		return err
	}

	if err := initMCP(ctx, client); err != nil {
		return err
	}

	c.confluence = client
	return nil
}

func (c *Client) startRovo(ctx context.Context) error {
	cfg := c.cfg.Rovo

	authHeader, err := c.rovoAuthHeader(ctx)
	if err != nil {
		return err
	}

	t, err := transport.NewStreamableHTTP(cfg.URL,
		transport.WithHTTPHeaders(map[string]string{
			"Authorization": authHeader,
		}),
	)
	if err != nil {
		return err
	}

	client := mcpgo.NewClient(t)

	if err := initMCP(ctx, client); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	c.rovo = client
	return nil
}

func (c *Client) rovoAuthHeader(_ context.Context) (string, error) {
	cfg := c.cfg.Rovo
	if cfg.Email == "" || cfg.APIToken == "" {
		return "", fmt.Errorf("rovo email and api_token must be set")
	}
	creds := base64.StdEncoding.EncodeToString([]byte(cfg.Email + ":" + cfg.APIToken))
	return "Basic " + creds, nil
}

func initMCP(ctx context.Context, client *mcpgo.Client) error {
	req := mcp.InitializeRequest{}
	req.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	req.Params.ClientInfo = mcp.Implementation{Name: "atlassian-client", Version: "1.0.0"}

	if _, err := client.Initialize(ctx, req); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	return nil
}

func buildEnv(m map[string]string) []string {
	env := make([]string, 0, len(m))
	for k, v := range m {
		env = append(env, fmt.Sprintf("%s=%s", strings.ToUpper(k), v))
	}
	return env
}

func extractText(resp *mcp.CallToolResult) string {
	if resp == nil {
		return ""
	}
	var parts []string
	for _, c := range resp.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			parts = append(parts, tc.Text)
		}
	}
	return strings.Join(parts, "\n")
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
