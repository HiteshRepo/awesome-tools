package atlassian

// Config holds the MCP server configuration for Atlassian services.
type Config struct {
	Jira       ServerConfig
	Confluence ServerConfig
	Rovo       RovoConfig
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
