package systemd

import (
	"fmt"

	"github.com/thedataflows/confedit/internal/types"
)

// Config represents the configuration for a systemd target
type Config struct {
	Unit       string                 `json:"unit"`
	Section    string                 `json:"section"`
	Properties map[string]interface{} `json:"properties"`
	Reload     bool                   `json:"reload,omitempty"`
}

// Type implements TargetConfig interface
func (c *Config) Type() string {
	return types.TYPE_SYSTEMD
}

// Validate checks if the systemd configuration is valid
func (c *Config) Validate() error {
	if c.Unit == "" {
		return fmt.Errorf("unit is required for systemd target")
	}
	if c.Section == "" {
		return fmt.Errorf("section is required for systemd target")
	}
	return nil
}

// Target is a type alias for systemd-based configuration targets
type Target = types.BaseTarget[*Config]

// NewTarget creates a new systemd configuration target
func NewTarget(name, unit, section string) *Target {
	return &Target{
		Name:     name,
		Type:     types.TYPE_SYSTEMD,
		Metadata: make(map[string]interface{}),
		Config: &Config{
			Unit:       unit,
			Section:    section,
			Properties: make(map[string]interface{}),
		},
	}
}
