package reconciler

import "github.com/thedataflows/confedit/internal/types"

// Reconciler is the interface for reconciliation engines
type Reconciler interface {
	// Validate checks if all targets are valid
	Validate(targets []types.AnyTarget) error

	// Reconcile applies changes to all targets
	Reconcile(targets []types.AnyTarget) error

	// Registry returns the feature registry for accessing executors
	Registry() interface{}
}

// Ensure implementation satisfies the interface
var _ Reconciler = (*ReconciliationEngine)(nil)
