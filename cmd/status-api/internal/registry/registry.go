// Package registry loads and validates the config/devices.yaml device list.
package registry

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Device kinds.
const (
	KindK8sNode = "k8s-node"
	KindMac     = "mac"
	KindLinux   = "linux"
	KindIOS     = "ios"
)

// Ownership values.
const (
	OwnershipPersonal = "personal"
	OwnershipCompany  = "company"
)

// Device represents one entry in config/devices.yaml.
type Device struct {
	ID                string `yaml:"id"`
	Name              string `yaml:"name"`
	Kind              string `yaml:"kind"`
	Ownership         string `yaml:"ownership"`
	Host              string `yaml:"host,omitempty"`
	Port              int    `yaml:"port,omitempty"`
	CleanupOnOffboard bool   `yaml:"cleanup_on_offboard,omitempty"`
}

type devicesFile struct {
	Devices []Device `yaml:"devices"`
}

// Load reads and validates the devices YAML file at path, returning the device list.
func Load(path string) ([]Device, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var f devicesFile
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	seen := map[string]bool{}
	for i, d := range f.Devices {
		if d.ID == "" {
			return nil, fmt.Errorf("device[%d]: missing required field 'id'", i)
		}
		if d.Name == "" {
			return nil, fmt.Errorf("device %q: missing required field 'name'", d.ID)
		}
		if d.Kind == "" {
			return nil, fmt.Errorf("device %q: missing required field 'kind'", d.ID)
		}
		switch d.Kind {
		case KindK8sNode, KindMac, KindLinux, KindIOS:
		default:
			return nil, fmt.Errorf("device %q: invalid kind %q (must be k8s-node, mac, linux, or ios)", d.ID, d.Kind)
		}
		if d.Ownership == "" {
			return nil, fmt.Errorf("device %q: missing required field 'ownership'", d.ID)
		}
		switch d.Ownership {
		case OwnershipPersonal, OwnershipCompany:
		default:
			return nil, fmt.Errorf("device %q: invalid ownership %q (must be personal or company)", d.ID, d.Ownership)
		}
		if (d.Kind == KindMac || d.Kind == KindLinux) && d.Host == "" {
			return nil, fmt.Errorf("device %q (kind=%s): missing required field 'host'", d.ID, d.Kind)
		}
		if (d.Kind == KindMac || d.Kind == KindLinux) && d.Port == 0 {
			return nil, fmt.Errorf("device %q (kind=%s): missing required field 'port'", d.ID, d.Kind)
		}
		if seen[d.ID] {
			return nil, fmt.Errorf("duplicate device id %q", d.ID)
		}
		seen[d.ID] = true
	}

	return f.Devices, nil
}
