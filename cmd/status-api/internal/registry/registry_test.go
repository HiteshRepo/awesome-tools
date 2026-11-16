package registry_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/HiteshRepo/awesome-tools/cmd/status-api/internal/registry"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "devices-*.yaml")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoadValid(t *testing.T) {
	path := writeTemp(t, `
devices:
  - id: legion
    name: Legion 5i
    kind: k8s-node
    ownership: personal
  - id: macbook-work
    name: MacBook Work
    kind: mac
    host: 192.168.1.101
    port: 9100
    ownership: company
    cleanup_on_offboard: true
  - id: iphone-15
    name: iPhone 15
    kind: ios
    ownership: personal
`)
	devices, err := registry.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devices) != 3 {
		t.Fatalf("expected 3 devices, got %d", len(devices))
	}
}

func TestLoadMissingID(t *testing.T) {
	path := writeTemp(t, `
devices:
  - name: No ID Device
    kind: ios
    ownership: personal
`)
	_, err := registry.Load(path)
	if err == nil {
		t.Fatal("expected error for missing id, got nil")
	}
}

func TestLoadMissingKind(t *testing.T) {
	path := writeTemp(t, `
devices:
  - id: test
    name: Test
    ownership: personal
`)
	_, err := registry.Load(path)
	if err == nil {
		t.Fatal("expected error for missing kind")
	}
}

func TestLoadInvalidKind(t *testing.T) {
	path := writeTemp(t, `
devices:
  - id: test
    name: Test
    kind: tablet
    ownership: personal
`)
	_, err := registry.Load(path)
	if err == nil {
		t.Fatal("expected error for invalid kind")
	}
}

func TestLoadDuplicateID(t *testing.T) {
	path := writeTemp(t, `
devices:
  - id: same
    name: First
    kind: ios
    ownership: personal
  - id: same
    name: Second
    kind: ios
    ownership: personal
`)
	_, err := registry.Load(path)
	if err == nil {
		t.Fatal("expected error for duplicate id")
	}
}

func TestLoadMissingHostForMac(t *testing.T) {
	path := writeTemp(t, `
devices:
  - id: mac1
    name: Mac One
    kind: mac
    port: 9100
    ownership: personal
`)
	_, err := registry.Load(path)
	if err == nil {
		t.Fatal("expected error for missing host")
	}
}

func TestLoadMissingPortForLinux(t *testing.T) {
	path := writeTemp(t, `
devices:
  - id: lin1
    name: Linux One
    kind: linux
    host: 192.168.1.50
    ownership: personal
`)
	_, err := registry.Load(path)
	if err == nil {
		t.Fatal("expected error for missing port")
	}
}

func TestLoadFileNotFound(t *testing.T) {
	_, err := registry.Load(filepath.Join(t.TempDir(), "nonexistent.yaml"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadInvalidOwnership(t *testing.T) {
	path := writeTemp(t, `
devices:
  - id: dev1
    name: Dev One
    kind: ios
    ownership: unknown
`)
	_, err := registry.Load(path)
	if err == nil {
		t.Fatal("expected error for invalid ownership")
	}
}

func TestLoadMissingName(t *testing.T) {
	path := writeTemp(t, `
devices:
  - id: dev1
    kind: ios
    ownership: personal
`)
	_, err := registry.Load(path)
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}
