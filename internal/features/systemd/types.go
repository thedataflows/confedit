package systemd

import (
	"fmt"

	"github.com/thedataflows/confedit/internal/types"
	"github.com/thedataflows/confedit/internal/utils"
)

// Config represents the configuration for a systemd target
type Config struct {
	Unit       string                 `json:"unit"`
	Section    string                 `json:"section"`
	Properties map[string]interface{} `json:"properties"`
	Backup     bool                   `json:"backup,omitempty"`
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

// Target is a type alias for systemd targets
type Target = types.BaseTarget[*Config]

// NewTarget creates a new systemd target
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

// MergeConfig merges systemd target configs using deep map merging
func MergeConfig(existing, newTarget *Config) error {
	if err := utils.DeepMerge(existing.Properties, newTarget.Properties); err != nil {
		return fmt.Errorf("merge properties: %w", err)
	}

	if newTarget.Unit != "" {
		existing.Unit = newTarget.Unit
	}
	if newTarget.Section != "" {
		existing.Section = newTarget.Section
	}
	if newTarget.Reload {
		existing.Reload = true
	}

	return nil
}
