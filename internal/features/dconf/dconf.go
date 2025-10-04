package dconf

import (
	"fmt"

	"github.com/thedataflows/confedit/internal/engine"
	"github.com/thedataflows/confedit/internal/features"
	"github.com/thedataflows/confedit/internal/types"
)

// Feature implements the features.Feature interface for dconf targets
type Feature struct {
	executor engine.Executor
}

// New creates a new dconf feature
func New() features.Feature {
	return &Feature{
		executor: NewExecutor(),
	}
}

// Type returns the feature type identifier
func (f *Feature) Type() string {
	return types.TYPE_DCONF
}

// Executor returns the executor implementation for this feature
func (f *Feature) Executor() engine.Executor {
	return f.executor
}

// NewTarget creates a new dconf target instance
func (f *Feature) NewTarget(name string, config interface{}) (types.AnyTarget, error) {
	dconfConfig, ok := config.(*Config)
	if !ok {
		return nil, fmt.Errorf("invalid config type for dconf target, expected *dconf.Config")
	}

	return &Target{
		Name:     name,
		Type:     types.TYPE_DCONF,
		Metadata: make(map[string]interface{}),
		Config:   dconfConfig,
	}, nil
}

// ValidateConfig validates the dconf-specific configuration
func (f *Feature) ValidateConfig(config interface{}) error {
	dconfConfig, ok := config.(*Config)
	if !ok {
		return fmt.Errorf("invalid config type for dconf target, expected *dconf.Config")
	}

	return dconfConfig.Validate()
}

// Verify that Feature implements the features.Feature interface at compile time
var _ features.Feature = (*Feature)(nil)
