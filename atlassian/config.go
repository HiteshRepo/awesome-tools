package atlassian

// Config holds the MCP server configuration for Atlassian services.
type Config struct {
	Jira          ServerConfig
	Confluence    ServerConfig
	Rovo          RovoConfig
	JiraREST      JiraRESTConfig      // for REST API fallback
	ClaudeCodeMCP ClaudeCodeMCPConfig // preferred: reuses Claude Code OAuth token
}

// ClaudeCodeMCPConfig connects to an HTTP MCP server using the OAuth token
// that Claude Code has already stored in ~/.claude/.credentials.json.
// No additional credentials are required beyond running:
//
//	claude mcp add <ServerName> --transport http <URL>
type ClaudeCodeMCPConfig struct {
	// ServerName must match the name used with "claude mcp add" (e.g. "atlassian-vdc-workspace").
	ServerName string
	// CloudID is the Atlassian site passed to tools that require it (e.g. "veeam-vdc.atlassian.net").
	CloudID string
	// CredentialsPath overrides the default ~/.claude/.credentials.json location.
	CredentialsPath string
}

// JiraRESTConfig holds credentials for direct Jira REST API access.
type JiraRESTConfig struct {
	BaseURL  string // e.g. https://yourorg.atlassian.net
	Email    string
	APIToken string
}

// ServerConfig defines the command, args, and env for a stdio MCP server.
type ServerConfig struct {
	Command string
	Args    []string
	Env     map[string]string
}

// RovoConfig holds credentials for the Atlassian official remote MCP.
type RovoConfig struct {
	URL      string
	Email    string
	APIToken string
	CloudID  string // Atlassian site UUID
}
