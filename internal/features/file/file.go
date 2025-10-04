package file

import (
	"fmt"

	"github.com/thedataflows/confedit/internal/engine"
	"github.com/thedataflows/confedit/internal/features"
	"github.com/thedataflows/confedit/internal/features/file/formats"
	"github.com/thedataflows/confedit/internal/features/file/formats/ini"
	jsonformat "github.com/thedataflows/confedit/internal/features/file/formats/json"
	"github.com/thedataflows/confedit/internal/features/file/formats/toml"
	"github.com/thedataflows/confedit/internal/features/file/formats/xml"
	"github.com/thedataflows/confedit/internal/features/file/formats/yaml"
	"github.com/thedataflows/confedit/internal/types"
)

// Feature implements the features.Feature interface for file configuration targets
type Feature struct {
	registry *formats.Registry
	executor engine.Executor
}

// New creates a new file feature with all supported formats registered
func New() features.Feature {
	registry := formats.NewRegistry()

	// Register all supported formats
	registry.Register("ini", ini.New())
	registry.Register("yaml", yaml.New())
	registry.Register("toml", toml.New())
	registry.Register("json", jsonformat.New())
	registry.Register("xml", xml.New())

	return &Feature{
		registry: registry,
		executor: NewExecutor(registry),
	}
}

// Type returns the feature type identifier
func (f *Feature) Type() string {
	return types.TYPE_FILE
}

// Executor returns the executor implementation for this feature
func (f *Feature) Executor() engine.Executor {
	return f.executor
}

// NewTarget creates a new file target instance
func (f *Feature) NewTarget(name string, config interface{}) (types.AnyTarget, error) {
	fileConfig, ok := config.(*Config)
	if !ok {
		return nil, fmt.Errorf("invalid config type for file target, expected *file.Config")
	}

	return &Target{
		Name:     name,
		Type:     types.TYPE_FILE,
		Metadata: make(map[string]interface{}),
		Config:   fileConfig,
	}, nil
}

// ValidateConfig validates the file-specific configuration
func (f *Feature) ValidateConfig(config interface{}) error {
	fileConfig, ok := config.(*Config)
	if !ok {
		return fmt.Errorf("invalid config type for file target, expected *file.Config")
	}

	return fileConfig.Validate()
}

// Verify that Feature implements the features.Feature interface at compile time
var _ features.Feature = (*Feature)(nil)
