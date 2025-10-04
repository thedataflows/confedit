package reconciler

import (
	"fmt"

	"github.com/thedataflows/confedit/internal/features"
	"github.com/thedataflows/confedit/internal/features/dconf"
	"github.com/thedataflows/confedit/internal/features/file"
	"github.com/thedataflows/confedit/internal/features/sed"
	"github.com/thedataflows/confedit/internal/features/systemd"
	"github.com/thedataflows/confedit/internal/state"
	"github.com/thedataflows/confedit/internal/types"
	"github.com/thedataflows/confedit/internal/utils"
	log "github.com/thedataflows/go-lib-log"
)

// ReconciliationEngine is the reconciler that uses the feature registry
type ReconciliationEngine struct {
	registry     *features.Registry
	stateManager *state.Manager
	dryRun       bool
}

// NewReconciliationEngine creates a new reconciler with feature registry support
func NewReconciliationEngine(registry *features.Registry, stateManager *state.Manager, dryRun bool) *ReconciliationEngine {
	return &ReconciliationEngine{
		registry:     registry,
		stateManager: stateManager,
		dryRun:       dryRun,
	}
}

// Validate validates all targets using their feature executors
func (r *ReconciliationEngine) Validate(targets []types.AnyTarget) error {
	for _, target := range targets {
		executor, err := r.registry.Executor(target.GetType())
		if err != nil {
			return fmt.Errorf("no executor found for target type '%s': %w", target.GetType(), err)
		}

		if err := executor.Validate(target); err != nil {
			return fmt.Errorf("validation failed for target '%s': %w", target.GetName(), err)
		}
	}

	log.Info("engine", "All targets validated successfully")
	return nil
}

// Reconcile reconciles all targets to their desired state
func (r *ReconciliationEngine) Reconcile(targets []types.AnyTarget) error {
	log.Info("engine", "Starting reconciliation process")

	for _, target := range targets {
		if err := r.reconcileTarget(target); err != nil {
			return fmt.Errorf("reconcile target '%s': %w", target.GetName(), err)
		}
	}

	log.Info("engine", "Reconciliation completed successfully")
	return nil
}

func (r *ReconciliationEngine) reconcileTarget(target types.AnyTarget) error {
	log.Debugf("engine", "Reconciling target: %s (type: %s)", target.GetName(), target.GetType())

	executor, err := r.registry.Executor(target.GetType())
	if err != nil {
		return fmt.Errorf("no executor found for target type '%s': %w", target.GetType(), err)
	}

	// Get current state from the system
	currentSystemState, err := executor.GetCurrentState(target)
	if err != nil {
		return fmt.Errorf("get current system state: %w", err)
	}

	// Get the appropriate content based on target type
	targetContent := r.getTargetContent(target)

	// Compute diff with desired state
	diff, err := r.stateManager.ComputeDiffWithCurrent(target.GetName(), targetContent, currentSystemState)
	if err != nil {
		return fmt.Errorf("compute diff: %w", err)
	}

	// Apply changes if needed
	if !diff.IsEmpty() {
		log.Debugf("engine", "Found changes for target '%s', applying...", target.GetName())

		if r.dryRun {
			colorSupport := utils.NewColorSupport()
			log.Infof("engine", "DRY RUN: Would apply changes to target '%s'", target.GetName())
			log.Debugf("engine", "Changes: %+v", diff.Changes)

			diffOutput := diff.FormatDiff(colorSupport)
			if diffOutput != "" {
				fmt.Printf("Would apply:\n%s\n", diffOutput)
			}
		} else {
			if err := executor.Apply(target, diff); err != nil {
				return fmt.Errorf("apply changes: %w", err)
			}

			log.Infof("engine", "Successfully applied changes to target '%s'", target.GetName())
		}
	} else {
		log.Debugf("engine", "No changes needed for target '%s'", target.GetName())
	}

	return nil
}

func (r *ReconciliationEngine) getTargetContent(target types.AnyTarget) map[string]interface{} {
	switch target.GetType() {
	case types.TYPE_FILE:
		if fileTarget, ok := target.(*file.Target); ok {
			return fileTarget.GetConfig().Content
		}
	case types.TYPE_DCONF:
		if dconfTarget, ok := target.(*dconf.Target); ok {
			return dconfTarget.GetConfig().Settings
		}
	case types.TYPE_SYSTEMD:
		if systemdTarget, ok := target.(*systemd.Target); ok {
			return systemdTarget.GetConfig().Properties
		}
	case types.TYPE_SED:
		if sedTarget, ok := target.(*sed.Target); ok {
			return map[string]interface{}{
				"commands": sedTarget.GetConfig().Commands,
				"path":     sedTarget.GetConfig().Path,
			}
		}
	}
	return make(map[string]interface{})
}

// Registry returns the feature registry
func (r *ReconciliationEngine) Registry() interface{} {
	return r.registry
}
