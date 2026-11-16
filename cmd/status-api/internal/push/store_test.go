package push_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/HiteshRepo/awesome-tools/cmd/status-api/internal/push"
	"github.com/HiteshRepo/awesome-tools/cmd/status-api/internal/registry"
)

func makeStore() *push.Store {
	devices := []registry.Device{
		{ID: "iphone-15", Name: "iPhone 15", Kind: registry.KindIOS, Ownership: registry.OwnershipPersonal},
		{ID: "ipad-pro", Name: "iPad Pro", Kind: registry.KindIOS, Ownership: registry.OwnershipPersonal},
		{ID: "legion", Name: "Legion", Kind: registry.KindK8sNode, Ownership: registry.OwnershipPersonal},
	}
	return push.NewStore(devices, 2*time.Hour)
}

func TestHandlePushValid204(t *testing.T) {
	s := makeStore()

	body := `{"id":"iphone-15","battery_percent":80,"charging":false,"disk_free_bytes":10000000000,"disk_total_bytes":128000000000,"wifi_name":"HomeWiFi"}`
	req := httptest.NewRequest(http.MethodPost, "/push", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	s.HandlePush(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandlePushUnknownID404(t *testing.T) {
	s := makeStore()

	body := `{"id":"unknown-device","battery_percent":80}`
	req := httptest.NewRequest(http.MethodPost, "/push", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	s.HandlePush(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestHandlePushMalformedJSON400(t *testing.T) {
	s := makeStore()

	req := httptest.NewRequest(http.MethodPost, "/push", bytes.NewBufferString("not json"))
	w := httptest.NewRecorder()
	s.HandlePush(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandlePushMissingID400(t *testing.T) {
	s := makeStore()

	body := `{"battery_percent":80}`
	req := httptest.NewRequest(http.MethodPost, "/push", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	s.HandlePush(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandlePushMethodNotAllowed(t *testing.T) {
	s := makeStore()
	req := httptest.NewRequest(http.MethodGet, "/push", nil)
	w := httptest.NewRecorder()
	s.HandlePush(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestPushAppearsInGet(t *testing.T) {
	s := makeStore()

	batt := float64(75)
	charging := true
	wifi := "HomeWiFi"
	body, _ := json.Marshal(push.IOSPayload{
		ID:             "iphone-15",
		BatteryPercent: &batt,
		Charging:       &charging,
		WiFiName:       &wifi,
	})
	req := httptest.NewRequest(http.MethodPost, "/push", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	s.HandlePush(w, req)

	p := s.Get("iphone-15")
	if p == nil {
		t.Fatal("expected stored payload, got nil")
	}
	if p.BatteryPercent == nil || *p.BatteryPercent != 75 {
		t.Errorf("expected battery_percent=75, got %v", p.BatteryPercent)
	}
}

func TestIsStaleWhenNoPayload(t *testing.T) {
	s := makeStore()
	if !s.IsStale("iphone-15") {
		t.Error("device with no push should be stale")
	}
}

func TestIsNotStaleAfterRecentPush(t *testing.T) {
	s := push.NewStore([]registry.Device{
		{ID: "iphone-15", Name: "iPhone 15", Kind: registry.KindIOS, Ownership: registry.OwnershipPersonal},
	}, 2*time.Hour)

	body := `{"id":"iphone-15","battery_percent":80}`
	req := httptest.NewRequest(http.MethodPost, "/push", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	s.HandlePush(w, req)

	if s.IsStale("iphone-15") {
		t.Error("device should not be stale immediately after push")
	}
}

func TestIsStaleAfterThreshold(t *testing.T) {
	// staleAfter = 1 nanosecond to force stale immediately
	s := push.NewStore([]registry.Device{
		{ID: "iphone-15", Name: "iPhone 15", Kind: registry.KindIOS, Ownership: registry.OwnershipPersonal},
	}, 1*time.Nanosecond)

	body := `{"id":"iphone-15","battery_percent":80}`
	req := httptest.NewRequest(http.MethodPost, "/push", bytes.NewBufferString(body))
	s.HandlePush(httptest.NewRecorder(), req)

	time.Sleep(5 * time.Millisecond)
	if !s.IsStale("iphone-15") {
		t.Error("device should be stale after threshold")
	}
}
