//go:build linux

package metrics

import "github.com/shirou/gopsutil/v4/sensors"

func collectCPUTemp(snap *Snapshot) {
	temps, err := sensors.SensorsTemperatures()
	if err != nil || len(temps) == 0 {
		return
	}
	for _, t := range temps {
		if t.Temperature > 0 {
			v := round2(t.Temperature)
			snap.CPUTempCelsius = &v
			return
		}
	}
}
