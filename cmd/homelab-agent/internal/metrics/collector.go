package metrics

import (
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
)

// Snapshot holds the current metrics for this host.
type Snapshot struct {
	CPUPercent     float64  `json:"cpu_percent"`
	MemPercent     float64  `json:"mem_percent"`
	DiskPercent    float64  `json:"disk_percent"`
	UptimeSeconds  uint64   `json:"uptime_seconds"`
	Timestamp      string   `json:"timestamp"`
	BatteryPercent *float64 `json:"battery_percent,omitempty"`
	Charging       *bool    `json:"charging,omitempty"`
	CPUTempCelsius *float64 `json:"cpu_temp_celsius,omitempty"`
}

// Collect gathers current metrics from the host.
func Collect() (Snapshot, error) {
	cpuPcts, err := cpu.Percent(500*time.Millisecond, false)
	if err != nil {
		return Snapshot{}, err
	}
	var cpuPct float64
	if len(cpuPcts) > 0 {
		cpuPct = cpuPcts[0]
	}

	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return Snapshot{}, err
	}

	diskStat, err := disk.Usage("/")
	if err != nil {
		return Snapshot{}, err
	}

	uptimeSec, err := host.Uptime()
	if err != nil {
		return Snapshot{}, err
	}

	snap := Snapshot{
		CPUPercent:    round2(cpuPct),
		MemPercent:    round2(vmStat.UsedPercent),
		DiskPercent:   round2(diskStat.UsedPercent),
		UptimeSeconds: uptimeSec,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
	}

	collectBattery(&snap)
	collectCPUTemp(&snap)

	return snap, nil
}

func round2(f float64) float64 {
	return float64(int(f*100)) / 100
}
