package cmd

import (
	"fmt"

	"github.com/alecthomas/kong"
	"github.com/thedataflows/confedit/internal/features"
	"github.com/thedataflows/confedit/internal/features/dconf"
	"github.com/thedataflows/confedit/internal/features/file"
	"github.com/thedataflows/confedit/internal/features/sed"
	"github.com/thedataflows/confedit/internal/features/systemd"
	"github.com/thedataflows/confedit/internal/reconciler"
	"github.com/thedataflows/confedit/internal/state"
	"github.com/thedataflows/confedit/internal/types"
	"github.com/thedataflows/confedit/internal/utils"
	log "github.com/thedataflows/go-lib-log"
)

// StatusCmd performs status checks on the target system
type StatusCmd struct {
	Targets []string `arg:"" optional:"" help:"Rest of the arguments are a list of target names"`
}

func (c *StatusCmd) Run(ctx *kong.Context, cli *CLI) error {
	dryRun := ""
	if cli.DryRun {
		dryRun = " (DRY RUN)"
	}
	log.Infof(PKG_CMD, "Starting status operation%s", dryRun)
	log.Debugf(PKG_CMD, "Status command options: %+v; context: %+v", cli, ctx.Args)

	// Initialize shared components - always use dry-run for status
	dryRunTrue := true
	cmdCtx, err := InitializeCommand(cli, c.Targets, &dryRunTrue, nil)
	if err != nil {
		return err
	}

	// Check status for each target
	hasChanges := false
	for _, target := range cmdCtx.Targets {
		changes, err := c.checkTargetStatus(target, cmdCtx.Reconciler, cmdCtx.StateManager)
		if err != nil {
			return fmt.Errorf("check status for target %s: %w", target.GetName(), err)
		}
		if changes {
			hasChanges = true
		}
	}

	// Summary
	if hasChanges {
		log.Warnf(PKG_CMD, "Drift detected - some targets need updates")
	} else {
		log.Infof(PKG_CMD, "All targets are in sync")
	}

	return nil
}

func (c *StatusCmd) checkTargetStatus(target types.AnyTarget, rec reconciler.Reconciler, stateManager *state.Manager) (bool, error) {
	// Initialize color support
	colorSupport := utils.NewColorSupport()

	log.Infof(PKG_CMD, "Checking status for target: %s (type: %s)", target.GetName(), target.GetType())

	// Get the executor for this target type from the reconciler's registry
	registry := rec.Registry().(*features.Registry)
	executor, err := registry.Executor(target.GetType())
	if err != nil {
		return false, fmt.Errorf("no executor found for target type '%s': %w", target.GetType(), err)
	}

	// Get current state from the system
	currentSystemState, err := executor.CurrentState(target)
	if err != nil {
		return false, fmt.Errorf("get current system state: %w", err)
	}

	// Get the appropriate content based on target type
	var targetContent map[string]interface{}
	switch target.GetType() {
	case types.TYPE_FILE:
		if fileTarget, ok := target.(*file.Target); ok {
			targetContent = fileTarget.GetConfig().Content
		}
	case types.TYPE_DCONF:
		if dconfTarget, ok := target.(*dconf.Target); ok {
			targetContent = dconfTarget.GetConfig().Settings
		}
	case types.TYPE_SYSTEMD:
		if systemdTarget, ok := target.(*systemd.Target); ok {
			targetContent = systemdTarget.GetConfig().Properties
		}
	case types.TYPE_SED:
		if sedTarget, ok := target.(*sed.Target); ok {
			// For sed targets, we check if the commands would result in changes
			// The current system state contains the file content
			targetContent = map[string]interface{}{
				"commands": sedTarget.GetConfig().Commands,
				"path":     sedTarget.GetConfig().Path,
			}
		}
	default:
		return false, fmt.Errorf("unsupported target type: %s", target.GetType())
	}

	// Compute diff to check for drift
	diff, err := stateManager.ComputeDiffWithCurrent(target.GetName(), targetContent, currentSystemState)
	if err != nil {
		return false, fmt.Errorf("compute diff: %w", err)
	}

	// Report status with colors
	if diff.IsEmpty() {
		fmt.Printf("%s %s: No changes needed\n", colorSupport.Green("✓"), target.GetName())
		return false, nil
	} else {
		fmt.Printf("%s %s: Changes required\n", colorSupport.Yellow("⚠"), target.GetName())

		// Show formatted diff with colors
		diffOutput := diff.FormatDiff(colorSupport)
		if diffOutput != "" {
			fmt.Println(diffOutput)
		}

		return true, nil
	}
}
