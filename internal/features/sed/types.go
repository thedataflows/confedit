package sed

import (
	"fmt"

	"github.com/thedataflows/confedit/internal/types"
)

// Config represents the configuration for a sed target
type Config struct {
	Path     string            `json:"path"`
	Commands []string          `json:"commands"`
	Backup   bool              `json:"backup,omitempty"`
	Options  map[string]string `json:"options,omitempty"`
}

// Type implements TargetConfig interface
func (c *Config) Type() string {
	return types.TYPE_SED
}

// Validate checks if the sed configuration is valid
func (c *Config) Validate() error {
	if c.Path == "" {
		return fmt.Errorf("path is required for sed target")
	}
	if len(c.Commands) == 0 {
		return fmt.Errorf("at least one sed command is required")
	}
	return nil
}

// Target is a type alias for sed targets
type Target = types.BaseTarget[*Config]

// NewTarget creates a new sed target
func NewTarget(name, path string, commands []string) *Target {
	return &Target{
		Name:     name,
		Type:     types.TYPE_SED,
		Metadata: make(map[string]interface{}),
		Config: &Config{
			Path:     path,
			Commands: commands,
		},
	}
}
