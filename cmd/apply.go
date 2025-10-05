package cmd

import (
	"fmt"

	"github.com/alecthomas/kong"
	log "github.com/thedataflows/go-lib-log"
)

// ApplyCmd performs edit operations on given files
type ApplyCmd struct {
	Targets []string `arg:"" optional:"" help:"Rest of the arguments are a list of target names"`
	Force   bool     `help:"Force apply even if validation fails"`
	Backup  bool     `help:"Create backup of files before modification" default:"true"`
}

func (c *ApplyCmd) Run(ctx *kong.Context, cli *CLI) error {
	dryRun := ""
	if cli.DryRun {
		dryRun = " (DRY RUN)"
	}
	log.Infof(PKG_CMD, "Starting apply operation%s", dryRun)
	log.Debugf(PKG_CMD, "Apply command options: %+v; context: %+v", cli, ctx.Args)

	// Initialize shared components
	cmdCtx, err := InitializeCommand(cli, c.Targets, nil, &c.Backup)
	if err != nil {
		return err
	}

	// Validate
	if err := cmdCtx.Reconciler.Validate(cmdCtx.Targets); err != nil && !c.Force {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Execute pre-apply hooks
	if cmdCtx.SystemConfig.Hooks != nil && len(cmdCtx.SystemConfig.Hooks.PreApply) > 0 {
		if err := cmdCtx.HookExecutor.ExecuteHooks(cmdCtx.SystemConfig.Hooks.PreApply, "pre_apply"); err != nil {
			return fmt.Errorf("pre-apply hooks failed: %w", err)
		}
	}

	// Reconcile and Apply
	if err := cmdCtx.Reconciler.Reconcile(cmdCtx.Targets); err != nil {
		return fmt.Errorf("reconciliation failed: %w", err)
	}

	// Execute post-apply hooks
	if cmdCtx.SystemConfig.Hooks != nil && len(cmdCtx.SystemConfig.Hooks.PostApply) > 0 {
		if err := cmdCtx.HookExecutor.ExecuteHooks(cmdCtx.SystemConfig.Hooks.PostApply, "post_apply"); err != nil {
			return fmt.Errorf("post-apply hooks failed: %w", err)
		}
	}

	return nil
}
