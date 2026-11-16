//go:build darwin

package metrics

import (
	"os/exec"
	"strconv"
	"strings"
)

func collectBattery(snap *Snapshot) {
	out, err := exec.Command("pmset", "-g", "batt").Output()
	if err != nil {
		return
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if !strings.Contains(line, "%") {
			continue
		}
		// Example line: "-InternalBattery-0 (id=12345678)\t80%; discharging; 2:30 remaining"
		parts := strings.Fields(line)
		for _, p := range parts {
			p = strings.TrimSuffix(p, ";")
			if strings.HasSuffix(p, "%") {
				pctStr := strings.TrimSuffix(p, "%")
				pct, err := strconv.ParseFloat(pctStr, 64)
				if err != nil {
					continue
				}
				snap.BatteryPercent = &pct

				charging := strings.Contains(line, "charging") && !strings.Contains(line, "discharging")
				snap.Charging = &charging
				return
			}
		}
	}
}
