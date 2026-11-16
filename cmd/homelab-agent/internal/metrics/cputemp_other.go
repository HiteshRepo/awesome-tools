//go:build !linux

package metrics

// collectCPUTemp is a no-op on non-Linux platforms.
func collectCPUTemp(_ *Snapshot) {}
