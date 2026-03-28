package atlassian

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// claudeCredential mirrors one entry from ~/.claude/.credentials.json → mcpOAuth.
type claudeCredential struct {
	ServerName   string `json:"serverName"`
	ServerURL    string `json:"serverUrl"`
	AccessToken  string `json:"accessToken"`
	ExpiresAt    int64  `json:"expiresAt"` // ms since epoch
	RefreshToken string `json:"refreshToken"`
	ClientID     string `json:"clientId"`
	Scope        string `json:"scope"`
}

// loadClaudeToken returns a valid Bearer token for the named MCP server.
// It reads ~/.claude/.credentials.json (or credPath if non-empty), finds the
// entry whose serverName matches, refreshes the access token if it has
// expired, writes the updated credential back to disk, and returns the token.
func loadClaudeToken(credPath, serverName string) (string, error) {
	if credPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("home dir: %w", err)
		}
		credPath = filepath.Join(home, ".claude", ".credentials.json")
	}

	raw, err := os.ReadFile(credPath)
	if err != nil {
		return "", fmt.Errorf("read credentials: %w", err)
	}

	var store struct {
		MCPOAuth map[string]claudeCredential `json:"mcpOAuth"`
	}
	if err := json.Unmarshal(raw, &store); err != nil {
		return "", fmt.Errorf("parse credentials: %w", err)
	}

	// Keys are "serverName|clientId" — match by serverName prefix.
	var key string
	var cred claudeCredential
	for k, v := range store.MCPOAuth {
		if v.ServerName == serverName || strings.HasPrefix(k, serverName+"|") {
			key = k
			cred = v
			break
		}
	}
	if key == "" {
		return "", fmt.Errorf("no credential found for MCP server %q — run: claude mcp add %s --transport http <url>", serverName, serverName)
	}

	// Refresh if expired (with a 30-second buffer).
	if time.Now().UnixMilli()+30_000 >= cred.ExpiresAt {
		if cred.RefreshToken == "" {
			return "", fmt.Errorf("access token expired and no refresh token available for %q", serverName)
		}
		refreshed, err := refreshOAuthToken(cred)
		if err != nil {
			return "", fmt.Errorf("refresh token: %w", err)
		}
		cred = refreshed
		store.MCPOAuth[key] = cred

		updated, err := json.Marshal(store)
		if err != nil {
			return "", fmt.Errorf("marshal credentials: %w", err)
		}
		if err := os.WriteFile(credPath, updated, 0600); err != nil {
			// Non-fatal: we still have the new token in memory.
			fmt.Fprintf(os.Stderr, "warning: could not persist refreshed token: %v\n", err)
		}
	}

	return cred.AccessToken, nil
}

// claudeCodeMCPServerURL returns the serverUrl stored for the named MCP server.
func claudeCodeMCPServerURL(credPath, serverName string) (string, error) {
	if credPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("home dir: %w", err)
		}
		credPath = filepath.Join(home, ".claude", ".credentials.json")
	}

	raw, err := os.ReadFile(credPath)
	if err != nil {
		return "", fmt.Errorf("read credentials: %w", err)
	}

	var store struct {
		MCPOAuth map[string]claudeCredential `json:"mcpOAuth"`
	}
	if err := json.Unmarshal(raw, &store); err != nil {
		return "", fmt.Errorf("parse credentials: %w", err)
	}

	for k, v := range store.MCPOAuth {
		if v.ServerName == serverName || strings.HasPrefix(k, serverName+"|") {
			return v.ServerURL, nil
		}
	}
	return "", fmt.Errorf("no server URL found for %q", serverName)
}

// refreshOAuthToken exchanges a refresh token for a new access token using
// the token endpoint derived from the server URL.
func refreshOAuthToken(cred claudeCredential) (claudeCredential, error) {
	// Derive base URL from server URL: strip /mcp suffix.
	baseURL := strings.TrimSuffix(cred.ServerURL, "/mcp")
	tokenEndpoint := baseURL + "/oauth/token"

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", cred.RefreshToken)
	form.Set("client_id", cred.ClientID)

	resp, err := http.PostForm(tokenEndpoint, form)
	if err != nil {
		return cred, fmt.Errorf("POST %s: %w", tokenEndpoint, err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return cred, fmt.Errorf("token endpoint returned %d: %s", resp.StatusCode, body)
	}

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"` // seconds
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return cred, fmt.Errorf("parse token response: %w", err)
	}

	cred.AccessToken = result.AccessToken
	if result.RefreshToken != "" {
		cred.RefreshToken = result.RefreshToken
	}
	if result.ExpiresIn > 0 {
		cred.ExpiresAt = time.Now().UnixMilli() + result.ExpiresIn*1000
	}

	return cred, nil
}
