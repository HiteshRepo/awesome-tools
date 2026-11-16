// Package push stores the most-recent push payload per iOS device id.
package push

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/HiteshRepo/awesome-tools/cmd/status-api/internal/registry"
)

// IOSPayload is the data an iOS Shortcut sends via POST /push.
// Only id is required; all metric fields are optional.
type IOSPayload struct {
	ID             string    `json:"id"`
	BatteryPercent *float64  `json:"battery_percent,omitempty"`
	Charging       *bool     `json:"charging,omitempty"`
	DiskFreeBytes  *int64    `json:"disk_free_bytes,omitempty"`
	DiskTotalBytes *int64    `json:"disk_total_bytes,omitempty"`
	WiFiName       *string   `json:"wifi_name,omitempty"`
	ReceivedAt     time.Time `json:"-"`
}

// Store holds the most-recent push payload per device id (in-memory, not persisted).
type Store struct {
	mu      sync.RWMutex
	entries map[string]*IOSPayload
	devices map[string]registry.Device
	staleAfter time.Duration
}

// NewStore creates a Store with the registered iOS devices and a stale threshold.
func NewStore(devices []registry.Device, staleAfter time.Duration) *Store {
	dm := map[string]registry.Device{}
	for _, d := range devices {
		if d.Kind == registry.KindIOS {
			dm[d.ID] = d
		}
	}
	return &Store{
		entries:    make(map[string]*IOSPayload),
		devices:    dm,
		staleAfter: staleAfter,
	}
}

// HandlePush is the HTTP handler for POST /push.
func (s *Store) HandlePush(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload IOSPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, fmt.Sprintf("invalid JSON: %v", err), http.StatusBadRequest)
		return
	}
	if payload.ID == "" {
		http.Error(w, "missing required field 'id'", http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	_, known := s.devices[payload.ID]
	s.mu.RUnlock()

	if !known {
		http.Error(w, fmt.Sprintf("unknown device id %q", payload.ID), http.StatusNotFound)
		return
	}

	payload.ReceivedAt = time.Now().UTC()

	s.mu.Lock()
	s.entries[payload.ID] = &payload
	s.mu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

// Get returns the most-recent payload for a device id (nil if never received).
func (s *Store) Get(id string) *IOSPayload {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.entries[id]
}

// IsStale returns true if no push has been received within staleAfter.
func (s *Store) IsStale(id string) bool {
	p := s.Get(id)
	if p == nil {
		return true
	}
	return time.Since(p.ReceivedAt) > s.staleAfter
}
