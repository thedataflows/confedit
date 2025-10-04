package cmd

import (
	"fmt"
	"slices"

	"github.com/thedataflows/confedit/internal/config"
	"github.com/thedataflows/confedit/internal/features"
	"github.com/thedataflows/confedit/internal/features/dconf"
	"github.com/thedataflows/confedit/internal/features/file"
	"github.com/thedataflows/confedit/internal/features/sed"
	"github.com/thedataflows/confedit/internal/features/systemd"
	"github.com/thedataflows/confedit/internal/reconciler"
	"github.com/thedataflows/confedit/internal/state"
	"github.com/thedataflows/confedit/internal/types"
)

// CommandContext holds the shared initialization components for apply and status commands
type CommandContext struct {
	StateManager *state.Manager
	Reconciler   reconciler.Reconciler
	Loader       *config.CueConfigLoader
	SystemConfig *types.SystemConfig
	Targets      []types.AnyTarget
	HookExecutor *reconciler.HookExecutor
}

// initializeFeatureRegistry creates and registers all available features
func initializeFeatureRegistry() *features.Registry {
	registry := features.NewRegistry()
	registry.Register(file.New())
	registry.Register(dconf.New())
	registry.Register(sed.New())
	registry.Register(systemd.New())
	return registry
}

// InitializeCommand performs common initialization for apply and status commands
func InitializeCommand(cli *CLI, currentTargets []string, dryRunOverride *bool, backupOverride *bool) (*CommandContext, error) {
	dryRun := cli.DryRun
	if dryRunOverride != nil {
		dryRun = *dryRunOverride
	}

	// Initialize components
	stateManager := state.NewManager(cli.StateDir)
	loader := config.NewCueConfigLoader(cli.Config, cli.Schema)
	hookExecutor := reconciler.NewHookExecutor(dryRun)

	// Create reconciler with feature registry
	registry := initializeFeatureRegistry()
	rec := reconciler.NewReconciliationEngine(registry, stateManager, dryRun)

	// Load configuration
	systemConfig, err := loader.LoadConfiguration()
	if err != nil {
		return nil, err
	}

	// Filter targets if specified
	targets := filterTargets(systemConfig.Targets, currentTargets)

	// Apply backup override if specified
	if backupOverride != nil && *backupOverride {
		for i := range targets {
			if targets[i].GetType() == types.TYPE_FILE {
				if fileTarget, ok := targets[i].(*file.Target); ok {
					fileTarget.GetConfig().Backup = true
				}
			}
		}
	}

	if len(targets) == 0 {
		if len(currentTargets) > 0 {
			return nil, fmt.Errorf("No targets found matching any of: %s", currentTargets)
		} else {
			return nil, fmt.Errorf("No targets configured")
		}
	}

	return &CommandContext{
		StateManager: stateManager,
		Reconciler:   rec,
		Loader:       loader,
		SystemConfig: systemConfig,
		Targets:      targets,
		HookExecutor: hookExecutor,
	}, nil
}

// filterTargets filters the targets based on the target names using slices.ContainsFunc
func filterTargets(allTargets []types.AnyTarget, currentTargets []string) []types.AnyTarget {
	if len(currentTargets) == 0 {
		return allTargets
	}

	// Use slices.ContainsFunc for efficient filtering
	return slices.DeleteFunc(slices.Clone(allTargets), func(target types.AnyTarget) bool {
		return !slices.Contains(currentTargets, target.GetName())
	})
}
