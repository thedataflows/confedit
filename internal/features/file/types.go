package file

import (
	"fmt"

	"github.com/thedataflows/confedit/internal/types"
)

// Config represents the configuration for a file target
type Config struct {
	Path    string                 `json:"path"`
	Format  string                 `json:"format"` // "ini" | "yaml" | "toml" | "json" | "xml"
	Owner   string                 `json:"owner,omitempty"`
	Group   string                 `json:"group,omitempty"`
	Mode    string                 `json:"mode,omitempty"`
	Backup  bool                   `json:"backup,omitempty"`
	Content map[string]interface{} `json:"content"`
	Options map[string]interface{} `json:"options,omitempty"` // Format-specific options
}

// Type implements TargetConfig interface
func (c *Config) Type() string {
	return types.TYPE_FILE
}

// Validate checks if the file configuration is valid
func (c *Config) Validate() error {
	if c.Path == "" {
		return fmt.Errorf("path is required for file target")
	}
	if c.Format == "" {
		return fmt.Errorf("format is required for file target")
	}

	// Validate format is supported
	supportedFormats := map[string]bool{
		"ini":  true,
		"yaml": true,
		"toml": true,
		"json": true,
		"xml":  true,
	}
	if !supportedFormats[c.Format] {
		return fmt.Errorf("unsupported format: %s (supported: ini, yaml, toml, json, xml)", c.Format)
	}

	return nil
}

// Target is a type alias for file-based configuration targets
type Target = types.BaseTarget[*Config]

// NewTarget creates a new file configuration target
func NewTarget(name, path, format string) *Target {
	return &Target{
		Name:     name,
		Type:     types.TYPE_FILE,
		Metadata: make(map[string]interface{}),
		Config: &Config{
			Path:    path,
			Format:  format,
			Content: make(map[string]interface{}),
			Options: make(map[string]interface{}),
		},
	}
}

// INIValue represents the structure for INI values according to schema
type INIValue struct {
	Value   interface{} `json:"value"`
	Comment string      `json:"comment,omitempty"`
}
