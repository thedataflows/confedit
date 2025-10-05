package engine

import (
	"github.com/thedataflows/confedit/internal/state"
	"github.com/thedataflows/confedit/internal/types"
)

// Executor defines the interface for target-specific executors
// Each feature (file, dconf, systemd, sed) implements this interface
type Executor interface {
	// Apply applies the changes to the target
	Apply(target types.AnyTarget, diff *state.ConfigDiff) error

	// Validate checks if the target definition is valid
	Validate(target types.AnyTarget) error

	// CurrentState retrieves the current state of the target
	CurrentState(target types.AnyTarget) (map[string]interface{}, error)
}
