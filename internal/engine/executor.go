package engine

import (
	"github.com/thedataflows/confedit/internal/state"
	"github.com/thedataflows/confedit/internal/types"
)

// Executor defines the interface for target-specific executors
// Each feature (file, dconf, systemd, sed) implements this interface
type Executor interface {
	// Apply applies the configuration changes to the target
	Apply(target types.AnyTarget, diff *state.ConfigDiff) error

	// Validate checks if the target configuration is valid
	Validate(target types.AnyTarget) error

	// GetCurrentState retrieves the current state of the target
	GetCurrentState(target types.AnyTarget) (map[string]interface{}, error)
}
