package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/HiteshRepo/awesome-tools/cmd/status-api/internal/api"
	"github.com/HiteshRepo/awesome-tools/cmd/status-api/internal/collector"
	"github.com/HiteshRepo/awesome-tools/cmd/status-api/internal/push"
	"github.com/HiteshRepo/awesome-tools/cmd/status-api/internal/registry"
)

var testDevices = []registry.Device{
	{ID: "legion", Name: "Legion 5i", Kind: registry.KindK8sNode, Ownership: registry.OwnershipPersonal},
	{ID: "macbook-work", Name: "MacBook Work", Kind: registry.KindMac, Host: "192.168.1.101", Port: 9100, Ownership: registry.OwnershipCompany},
	{ID: "iphone-15", Name: "iPhone 15", Kind: registry.KindIOS, Ownership: registry.OwnershipPersonal},
}

func makeHandler() *api.Handler {
	agentPoll := collector.NewAgentPoller(testDevices, 60*time.Second, 5*time.Second)
	k8sColl := collector.NewK8sCollector(60 * time.Second)
	pushStore := push.NewStore(testDevices, 2*time.Hour)
	return api.New(testDevices, agentPoll, k8sColl, pushStore, 2*time.Hour)
}

func TestHandleHealthReturns200(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	api.HandleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body map[string]interface{}
	json.NewDecoder(w.Body).Decode(&body)
	if body["ok"] != true {
		t.Error("expected {\"ok\":true}")
	}
}

func TestHandleStatusReturnsOneEntryPerDevice(t *testing.T) {
	h := makeHandler()
	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	w := httptest.NewRecorder()
	h.HandleStatus(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var entries []map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&entries); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(entries) != len(testDevices) {
		t.Errorf("expected %d entries, got %d", len(testDevices), len(entries))
	}
}

func TestHandleStatusRequiredFields(t *testing.T) {
	h := makeHandler()
	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	w := httptest.NewRecorder()
	h.HandleStatus(w, req)

	var entries []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&entries)

	for _, e := range entries {
		for _, field := range []string{"id", "name", "kind", "ownership", "status"} {
			if _, ok := e[field]; !ok {
				t.Errorf("missing field %q in entry %v", field, e["id"])
			}
		}
	}
}

func TestHandleStatusMacAgentErrorStatus(t *testing.T) {
	h := makeHandler()
	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	w := httptest.NewRecorder()
	h.HandleStatus(w, req)

	var entries []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&entries)

	for _, e := range entries {
		if e["id"] == "macbook-work" {
			// No agent polled yet → error
			if e["status"] != "error" {
				t.Errorf("expected macbook-work status=error (no poll yet), got %v", e["status"])
			}
			return
		}
	}
	t.Error("macbook-work device not found in response")
}

func TestHandleStatusIOSStaleWhenNoPush(t *testing.T) {
	h := makeHandler()
	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	w := httptest.NewRecorder()
	h.HandleStatus(w, req)

	var entries []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&entries)

	for _, e := range entries {
		if e["id"] == "iphone-15" {
			if e["status"] != "stale" {
				t.Errorf("expected iphone-15 status=stale (no push), got %v", e["status"])
			}
			return
		}
	}
	t.Error("iphone-15 device not found in response")
}
