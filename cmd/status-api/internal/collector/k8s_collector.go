package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

// K8sMetrics holds the aggregated cluster metrics for the k8s-node device.
type K8sMetrics struct {
	CPUPercent    *float64 `json:"cpu_percent,omitempty"`
	MemPercent    *float64 `json:"mem_percent,omitempty"`
	PodCount      *int     `json:"pod_count,omitempty"`
	ErrorPodCount *int     `json:"error_pod_count,omitempty"`
	ArgoCDSync    *string  `json:"argocd_sync_status,omitempty"`
	OllamaModel   *string  `json:"ollama_active_model,omitempty"`
	OllamaVRAM    *int64   `json:"ollama_vram_used_bytes,omitempty"`
	KopiaLastRun  *string  `json:"kopia_last_backup,omitempty"`
	UpdatedAt     string   `json:"updated_at"`
}

// K8sCollector collects in-cluster metrics on an interval.
type K8sCollector struct {
	mu     sync.RWMutex
	latest *K8sMetrics

	interval      time.Duration
	k8sAPIBase    string
	argoCDAPIBase string
	ollamaAPIBase string
	kopiaAPIBase  string
	httpClient    *http.Client
}

// NewK8sCollector returns a collector using in-cluster defaults or env-overrides.
func NewK8sCollector(interval time.Duration) *K8sCollector {
	k8sBase := envOr("K8S_API_BASE", "https://kubernetes.default.svc")
	argoCDBase := envOr("ARGOCD_API_BASE", "http://argocd-server.argocd.svc.cluster.local")
	ollamaBase := envOr("OLLAMA_API_BASE", "http://ollama.ollama.svc.cluster.local:11434")
	kopiaBase := envOr("KOPIA_API_BASE", "http://kopia.kopia.svc.cluster.local:51515")

	return &K8sCollector{
		interval:      interval,
		k8sAPIBase:    k8sBase,
		argoCDAPIBase: argoCDBase,
		ollamaAPIBase: ollamaBase,
		kopiaAPIBase:  kopiaBase,
		httpClient:    &http.Client{Timeout: 10 * time.Second},
	}
}

// Start runs the collection loop until ctx is cancelled.
func (c *K8sCollector) Start(ctx context.Context) {
	c.collect()
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.collect()
		}
	}
}

// Get returns a copy of the latest collected metrics.
func (c *K8sCollector) Get() *K8sMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.latest == nil {
		return nil
	}
	cp := *c.latest
	return &cp
}

func (c *K8sCollector) collect() {
	m := &K8sMetrics{UpdatedAt: time.Now().UTC().Format(time.RFC3339)}

	// Pod counts via K8s API (uses in-cluster service-account token)
	c.collectPods(m)
	c.collectArgoCDSync(m)
	c.collectOllama(m)
	c.collectKopia(m)

	c.mu.Lock()
	c.latest = m
	c.mu.Unlock()
}

func (c *K8sCollector) collectPods(m *K8sMetrics) {
	token, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		return
	}

	req, _ := http.NewRequest(http.MethodGet, c.k8sAPIBase+"/api/v1/pods", nil)
	req.Header.Set("Authorization", "Bearer "+string(token))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var body struct {
		Items []struct {
			Status struct {
				Phase string `json:"phase"`
			} `json:"status"`
		} `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return
	}

	total := len(body.Items)
	errCount := 0
	for _, item := range body.Items {
		switch item.Status.Phase {
		case "Failed", "Unknown":
			errCount++
		}
	}
	m.PodCount = &total
	m.ErrorPodCount = &errCount
}

func (c *K8sCollector) collectArgoCDSync(m *K8sMetrics) {
	url := fmt.Sprintf("%s/api/v1/applications", c.argoCDAPIBase)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var body struct {
		Items []struct {
			Status struct {
				Sync struct {
					Status string `json:"status"`
				} `json:"sync"`
			} `json:"status"`
		} `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return
	}

	status := "Synced"
	for _, app := range body.Items {
		if app.Status.Sync.Status != "Synced" {
			status = "OutOfSync"
			break
		}
	}
	m.ArgoCDSync = &status
}

func (c *K8sCollector) collectOllama(m *K8sMetrics) {
	resp, err := c.httpClient.Get(c.ollamaAPIBase + "/api/ps")
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var body struct {
		Models []struct {
			Name  string `json:"name"`
			SizeVRAM int64 `json:"size_vram"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return
	}

	if len(body.Models) > 0 {
		m.OllamaModel = &body.Models[0].Name
		m.OllamaVRAM = &body.Models[0].SizeVRAM
	}
}

func (c *K8sCollector) collectKopia(m *K8sMetrics) {
	resp, err := c.httpClient.Get(c.kopiaAPIBase + "/api/v1/snapshots")
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil || len(body) == 0 {
		return
	}

	// Extract startTime from the last snapshot entry
	var result struct {
		Snapshots []struct {
			StartTime string `json:"startTime"`
		} `json:"snapshots"`
	}
	if err := json.Unmarshal(body, &result); err != nil || len(result.Snapshots) == 0 {
		return
	}
	last := result.Snapshots[len(result.Snapshots)-1].StartTime
	m.KopiaLastRun = &last
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
