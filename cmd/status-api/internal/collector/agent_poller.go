// Package collector polls homelab-agents and k8s APIs, storing results.
package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/HiteshRepo/awesome-tools/cmd/status-api/internal/registry"
)

// AgentMetrics is the payload returned by homelab-agent GET /metrics.
type AgentMetrics struct {
	CPUPercent     float64  `json:"cpu_percent"`
	MemPercent     float64  `json:"mem_percent"`
	DiskPercent    float64  `json:"disk_percent"`
	UptimeSeconds  uint64   `json:"uptime_seconds"`
	Timestamp      string   `json:"timestamp"`
	BatteryPercent *float64 `json:"battery_percent,omitempty"`
	Charging       *bool    `json:"charging,omitempty"`
	CPUTempCelsius *float64 `json:"cpu_temp_celsius,omitempty"`
}

// AgentState holds the latest data for a mac/linux device.
type AgentState struct {
	Metrics  *AgentMetrics
	LastSeen *time.Time
	Error    bool
}

// AgentPoller polls homelab-agents for all mac/linux devices.
type AgentPoller struct {
	mu       sync.RWMutex
	states   map[string]*AgentState
	devices  []registry.Device
	interval time.Duration
	timeout  time.Duration
	client   *http.Client
}

// NewAgentPoller creates a poller with the given interval and poll timeout.
func NewAgentPoller(devices []registry.Device, interval, timeout time.Duration) *AgentPoller {
	p := &AgentPoller{
		states:   make(map[string]*AgentState),
		devices:  devices,
		interval: interval,
		timeout:  timeout,
		client:   &http.Client{Timeout: timeout},
	}
	for _, d := range devices {
		if d.Kind == registry.KindMac || d.Kind == registry.KindLinux {
			p.states[d.ID] = &AgentState{}
		}
	}
	return p
}

// Start launches polling goroutines; runs until ctx is cancelled.
func (p *AgentPoller) Start(ctx context.Context) {
	for _, d := range p.devices {
		if d.Kind != registry.KindMac && d.Kind != registry.KindLinux {
			continue
		}
		go p.pollLoop(ctx, d)
	}
}

func (p *AgentPoller) pollLoop(ctx context.Context, d registry.Device) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	// Poll immediately on start
	p.poll(d)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.poll(d)
		}
	}
}

func (p *AgentPoller) poll(d registry.Device) {
	url := fmt.Sprintf("http://%s:%d/metrics", d.Host, d.Port)
	resp, err := p.client.Get(url)
	if err != nil {
		p.mu.Lock()
		p.states[d.ID].Error = true
		p.mu.Unlock()
		return
	}
	defer resp.Body.Close()

	var m AgentMetrics
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		p.mu.Lock()
		p.states[d.ID].Error = true
		p.mu.Unlock()
		return
	}

	now := time.Now().UTC()
	p.mu.Lock()
	p.states[d.ID] = &AgentState{Metrics: &m, LastSeen: &now, Error: false}
	p.mu.Unlock()
}

// Get returns the current state for a device id.
func (p *AgentPoller) Get(id string) *AgentState {
	p.mu.RLock()
	defer p.mu.RUnlock()
	s, ok := p.states[id]
	if !ok {
		return nil
	}
	cp := *s
	return &cp
}
