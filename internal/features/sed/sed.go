package sed

import (
	"fmt"

	"github.com/thedataflows/confedit/internal/engine"
	"github.com/thedataflows/confedit/internal/features"
	"github.com/thedataflows/confedit/internal/types"
)

// Feature implements the features.Feature interface for sed targets
type Feature struct {
	executor engine.Executor
}

// New creates a new sed feature
func New() features.Feature {
	return &Feature{
		executor: NewExecutor(),
	}
}

// Type returns the feature type identifier
func (f *Feature) Type() string {
	return types.TYPE_SED
}

// Executor returns the executor implementation for this feature
func (f *Feature) Executor() engine.Executor {
	return f.executor
}

// NewTarget creates a new sed target instance
func (f *Feature) NewTarget(name string, config interface{}) (types.AnyTarget, error) {
	sedConfig, ok := config.(*Config)
	if !ok {
		return nil, fmt.Errorf("invalid config type for sed target, expected *sed.Config")
	}

	return &Target{
		Name:     name,
		Type:     types.TYPE_SED,
		Metadata: make(map[string]interface{}),
		Config:   sedConfig,
	}, nil
}

// ValidateConfig validates the sed-specific configuration
func (f *Feature) ValidateConfig(config interface{}) error {
	sedConfig, ok := config.(*Config)
	if !ok {
		return fmt.Errorf("invalid config type for sed target, expected *sed.Config")
	}

	return sedConfig.Validate()
}

// Verify that Feature implements the features.Feature interface at compile time
var _ features.Feature = (*Feature)(nil)
