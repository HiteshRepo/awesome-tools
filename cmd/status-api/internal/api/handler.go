// Package api provides the HTTP handlers for status-api.
package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/HiteshRepo/awesome-tools/cmd/status-api/internal/collector"
	"github.com/HiteshRepo/awesome-tools/cmd/status-api/internal/push"
	"github.com/HiteshRepo/awesome-tools/cmd/status-api/internal/registry"
)

// Health values for device entries.
const (
	HealthHealthy = "healthy"
	HealthWarning = "warning"
	HealthError   = "error"
	HealthStale   = "stale"
)

// DeviceStatus is one entry in the GET /status response.
type DeviceStatus struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Kind      string      `json:"kind"`
	Ownership string      `json:"ownership"`
	Status    string      `json:"status"`
	LastSeen  *string     `json:"last_seen,omitempty"`
	Metrics   interface{} `json:"metrics,omitempty"`
}

// Handler holds all dependencies needed to serve the status API.
type Handler struct {
	devices    []registry.Device
	agentPoll  *collector.AgentPoller
	k8sColl    *collector.K8sCollector
	pushStore  *push.Store
	staleAfter time.Duration
}

// New creates a Handler.
func New(
	devices []registry.Device,
	agentPoll *collector.AgentPoller,
	k8sColl *collector.K8sCollector,
	pushStore *push.Store,
	staleAfter time.Duration,
) *Handler {
	return &Handler{
		devices:    devices,
		agentPoll:  agentPoll,
		k8sColl:    k8sColl,
		pushStore:  pushStore,
		staleAfter: staleAfter,
	}
}

// HandleHealth serves GET /health.
func HandleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"ok":true}`))
}

// HandleStatus serves GET /status.
func (h *Handler) HandleStatus(w http.ResponseWriter, _ *http.Request) {
	entries := make([]DeviceStatus, 0, len(h.devices))

	for _, d := range h.devices {
		entry := h.buildEntry(d)
		entries = append(entries, entry)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

func (h *Handler) buildEntry(d registry.Device) DeviceStatus {
	ds := DeviceStatus{
		ID:        d.ID,
		Name:      d.Name,
		Kind:      d.Kind,
		Ownership: d.Ownership,
		Status:    HealthHealthy,
	}

	switch d.Kind {
	case registry.KindK8sNode:
		m := h.k8sColl.Get()
		if m != nil {
			ds.Metrics = m
		}

	case registry.KindMac, registry.KindLinux:
		state := h.agentPoll.Get(d.ID)
		if state == nil || state.Error || state.Metrics == nil {
			ds.Status = HealthError
		} else {
			ds.Metrics = state.Metrics
			if state.LastSeen != nil {
				t := state.LastSeen.Format(time.RFC3339)
				ds.LastSeen = &t
			}
		}

	case registry.KindIOS:
		p := h.pushStore.Get(d.ID)
		if h.pushStore.IsStale(d.ID) {
			ds.Status = HealthStale
		}
		if p != nil {
			t := p.ReceivedAt.Format(time.RFC3339)
			ds.LastSeen = &t
			ds.Metrics = p
		}
	}

	return ds
}
