//go:build !darwin && !linux

package metrics

// collectBattery is a no-op on unsupported platforms.
func collectBattery(_ *Snapshot) {}
