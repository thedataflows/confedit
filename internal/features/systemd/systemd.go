package systemd

import (
	"fmt"

	"github.com/thedataflows/confedit/internal/engine"
	"github.com/thedataflows/confedit/internal/features"
	"github.com/thedataflows/confedit/internal/types"
)

// Feature implements the features.Feature interface for systemd targets
type Feature struct {
	executor engine.Executor
}

// New creates a new systemd feature
func New() features.Feature {
	return &Feature{
		executor: NewExecutor(),
	}
}

// Type returns the feature type identifier
func (f *Feature) Type() string {
	return types.TYPE_SYSTEMD
}

// Executor returns the executor implementation for this feature
func (f *Feature) Executor() engine.Executor {
	return f.executor
}

// NewTarget creates a new systemd target instance
func (f *Feature) NewTarget(name string, config interface{}) (types.AnyTarget, error) {
	systemdConfig, ok := config.(*Config)
	if !ok {
		return nil, fmt.Errorf("invalid config type for systemd target, expected *systemd.Config")
	}

	return &Target{
		Name:     name,
		Type:     types.TYPE_SYSTEMD,
		Metadata: make(map[string]interface{}),
		Config:   systemdConfig,
	}, nil
}

// Validate validates the systemd target
func (f *Feature) Validate(config interface{}) error {
	systemdConfig, ok := config.(*Config)
	if !ok {
		return fmt.Errorf("invalid config type for systemd target, expected *systemd.Config")
	}

	return systemdConfig.Validate()
}

// Verify that Feature implements the features.Feature interface at compile time
var _ features.Feature = (*Feature)(nil)
