package features

import (
	"github.com/thedataflows/confedit/internal/engine"
	"github.com/thedataflows/confedit/internal/types"
)

// Feature represents a target feature (file, dconf, systemd, sed)
// Each feature is a self-contained module that handles a specific type of target
type Feature interface {
	// Type returns the feature type identifier (e.g., "file", "dconf", "systemd", "sed")
	Type() string

	// Executor returns the executor implementation for this feature
	Executor() engine.Executor

	// NewTarget creates a new target instance for this feature type
	NewTarget(name string, config interface{}) (types.AnyTarget, error)

	// ValidateConfig validates the feature-specific target definition
	Validate(config interface{}) error
}
