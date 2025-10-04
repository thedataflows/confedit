package dconf

import (
	"fmt"

	"github.com/thedataflows/confedit/internal/types"
)

// Config represents the configuration for a dconf target
type Config struct {
	User     string                 `json:"user,omitempty"`
	Schema   string                 `json:"schema"`
	Settings map[string]interface{} `json:"settings"`
}

// Type implements TargetConfig interface
func (c *Config) Type() string {
	return types.TYPE_DCONF
}

// Validate checks if the dconf configuration is valid
func (c *Config) Validate() error {
	if c.Schema == "" {
		return fmt.Errorf("schema is required for dconf target")
	}
	return nil
}

// Target is a type alias for dconf-based configuration targets
type Target = types.BaseTarget[*Config]

// NewTarget creates a new dconf configuration target
func NewTarget(name, schema string) *Target {
	return &Target{
		Name:     name,
		Type:     types.TYPE_DCONF,
		Metadata: make(map[string]interface{}),
		Config: &Config{
			Schema:   schema,
			Settings: make(map[string]interface{}),
		},
	}
}
