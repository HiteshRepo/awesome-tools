package metrics

import (
	"encoding/json"
	"testing"
	"time"
)

func TestSnapshotSerialisesAllRequiredFields(t *testing.T) {
	snap, err := Collect()
	if err != nil {
		t.Fatalf("Collect() error: %v", err)
	}

	data, err := json.Marshal(snap)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}

	var out map[string]interface{}
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}

	required := []string{"cpu_percent", "mem_percent", "disk_percent", "uptime_seconds", "timestamp"}
	for _, field := range required {
		if _, ok := out[field]; !ok {
			t.Errorf("missing required field %q in serialised snapshot", field)
		}
	}
}

func TestSnapshotTimestampIsRFC3339(t *testing.T) {
	snap, err := Collect()
	if err != nil {
		t.Fatalf("Collect() error: %v", err)
	}
	if _, err := time.Parse(time.RFC3339, snap.Timestamp); err != nil {
		t.Errorf("Timestamp %q is not RFC 3339: %v", snap.Timestamp, err)
	}
}

func TestSnapshotCPUPercentInRange(t *testing.T) {
	snap, err := Collect()
	if err != nil {
		t.Fatalf("Collect() error: %v", err)
	}
	if snap.CPUPercent < 0 || snap.CPUPercent > 100 {
		t.Errorf("CPUPercent out of range: %v", snap.CPUPercent)
	}
}

func TestSnapshotMemPercentInRange(t *testing.T) {
	snap, err := Collect()
	if err != nil {
		t.Fatalf("Collect() error: %v", err)
	}
	if snap.MemPercent < 0 || snap.MemPercent > 100 {
		t.Errorf("MemPercent out of range: %v", snap.MemPercent)
	}
}

func TestSnapshotDiskPercentInRange(t *testing.T) {
	snap, err := Collect()
	if err != nil {
		t.Fatalf("Collect() error: %v", err)
	}
	if snap.DiskPercent < 0 || snap.DiskPercent > 100 {
		t.Errorf("DiskPercent out of range: %v", snap.DiskPercent)
	}
}

func TestSnapshotUptimeIsPositive(t *testing.T) {
	snap, err := Collect()
	if err != nil {
		t.Fatalf("Collect() error: %v", err)
	}
	if snap.UptimeSeconds == 0 {
		t.Error("UptimeSeconds should be > 0 on a running system")
	}
}

func TestSnapshotOmitsNilBatteryFields(t *testing.T) {
	snap := Snapshot{
		CPUPercent:  10,
		MemPercent:  50,
		DiskPercent: 30,
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
	}
	// BatteryPercent and Charging are nil — they should be omitted from JSON
	data, _ := json.Marshal(snap)
	var out map[string]interface{}
	json.Unmarshal(data, &out)

	if _, ok := out["battery_percent"]; ok {
		t.Error("battery_percent should be omitted when nil")
	}
	if _, ok := out["charging"]; ok {
		t.Error("charging should be omitted when nil")
	}
}

func TestRound2(t *testing.T) {
	cases := []struct {
		in   float64
		want float64
	}{
		{0.0, 0.0},
		{100.0, 100.0},
		{33.3333, 33.33},
		{66.666, 66.66},
		{99.999, 99.99},
	}
	for _, c := range cases {
		got := round2(c.in)
		if got != c.want {
			t.Errorf("round2(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestSnapshotIncludesBatteryWhenPresent(t *testing.T) {
	pct := 85.5
	charging := true
	snap := Snapshot{
		CPUPercent:     10,
		MemPercent:     50,
		DiskPercent:    30,
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
		BatteryPercent: &pct,
		Charging:       &charging,
	}
	data, _ := json.Marshal(snap)
	var out map[string]interface{}
	json.Unmarshal(data, &out)

	if _, ok := out["battery_percent"]; !ok {
		t.Error("battery_percent should be present when set")
	}
	if _, ok := out["charging"]; !ok {
		t.Error("charging should be present when set")
	}
}
