package reconciler

import (
	"context"
	"fmt"
	"os"
	"strings"

	log "github.com/thedataflows/go-lib-log"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

// HookExecutor executes shell commands using mvdan.cc/sh for portability
type HookExecutor struct {
	dryRun bool
}

func NewHookExecutor(dryRun bool) *HookExecutor {
	return &HookExecutor{
		dryRun: dryRun,
	}
}

// ExecuteHooks executes a list of shell scripts
func (he *HookExecutor) ExecuteHooks(hooks []string, hookType string) error {
	if len(hooks) == 0 {
		return nil
	}

	log.Infof("engine", "Executing %s hooks (%d scripts)", hookType, len(hooks))

	for i, script := range hooks {
		if err := he.executeScript(script, fmt.Sprintf("%s[%d]", hookType, i)); err != nil {
			return fmt.Errorf("execute %s hook %d: %w", hookType, i, err)
		}
	}

	log.Infof("engine", "Successfully executed all %s hooks", hookType)
	return nil
}

// executeScript executes a full shell script using mvdan.cc/sh
func (he *HookExecutor) executeScript(script, identifier string) error {
	// Clean up the script - remove leading/trailing whitespace
	script = strings.TrimSpace(script)

	if script == "" {
		log.Debugf("engine", "Skipping empty hook %s", identifier)
		return nil
	}

	if he.dryRun {
		log.Infof("engine", "DRY RUN: Would execute hook %s:\n%s", identifier, script)
		return nil
	} else {
		log.Debugf("engine", "Executing hook %s:\n%s", identifier, script)
	}

	// Parse the shell script
	parser := syntax.NewParser()
	prog, err := parser.Parse(strings.NewReader(script), identifier)
	if err != nil {
		return fmt.Errorf("parse script: %w", err)
	}

	// Create interpreter with standard environment
	runner, err := interp.New(
		interp.StdIO(os.Stdin, os.Stdout, os.Stderr),
		interp.Env(expand.ListEnviron(os.Environ()...)),
	)
	if err != nil {
		return fmt.Errorf("create shell interpreter: %w", err)
	}

	// Execute the command
	ctx := context.Background()
	if err := runner.Run(ctx, prog); err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}

	log.Debugf("engine", "Successfully executed hook %s", identifier)
	return nil
}
