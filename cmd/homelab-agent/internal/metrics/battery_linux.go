//go:build linux

package metrics

import (
	"os"
	"strconv"
	"strings"
)

const (
	battCapacityPath = "/sys/class/power_supply/BAT0/capacity"
	battStatusPath   = "/sys/class/power_supply/BAT0/status"
)

func collectBattery(snap *Snapshot) {
	capBytes, err := os.ReadFile(battCapacityPath)
	if err != nil {
		return
	}
	pct, err := strconv.ParseFloat(strings.TrimSpace(string(capBytes)), 64)
	if err != nil {
		return
	}
	snap.BatteryPercent = &pct

	statusBytes, err := os.ReadFile(battStatusPath)
	if err != nil {
		charging := false
		snap.Charging = &charging
		return
	}
	charging := strings.TrimSpace(string(statusBytes)) == "Charging"
	snap.Charging = &charging
}
