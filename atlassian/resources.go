package atlassian

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
)

const accessibleResourcesURL = "https://api.atlassian.com/oauth/token/accessible-resources"

// AtlassianSite represents a single Atlassian Cloud site accessible with the given credentials.
type AtlassianSite struct {
	ID     string   `json:"id"`   // UUID — use this as cloud_id
	Name   string   `json:"name"` // human-readable org name
	URL    string   `json:"url"`  // e.g. https://myorg.atlassian.net
	Scopes []string `json:"scopes"`
}

// GetAccessibleResources calls the Atlassian API to list all Cloud sites accessible
// with the given email + API token. The returned site IDs are valid CloudID values.
func GetAccessibleResources(email, apiToken string) ([]AtlassianSite, error) {
	req, err := http.NewRequest(http.MethodGet, accessibleResourcesURL, nil)
	if err != nil {
		return nil, err
	}

	creds := base64.StdEncoding.EncodeToString([]byte(email + ":" + apiToken))
	req.Header.Set("Authorization", "Basic "+creds)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("invalid credentials (401) — check email and API token")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var sites []AtlassianSite
	if err := json.NewDecoder(resp.Body).Decode(&sites); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return sites, nil
}
