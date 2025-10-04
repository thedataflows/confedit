package cmd

import (
	"testing"

	"github.com/thedataflows/confedit/internal/features/file"
	"github.com/thedataflows/confedit/internal/types"
)

// TestInitializeCommand verifies the reconciler initialization
func TestInitializeCommand(t *testing.T) {
	cli := &CLI{
		Globals: Globals{
			Config:   "../testdata/01-example-config.cue",
			StateDir: "",
			DryRun:   true,
		},
	}

	ctx, err := InitializeCommand(cli, nil, nil, nil)
	if err != nil {
		t.Fatalf("InitializeCommand failed: %v", err)
	}

	if ctx.Reconciler == nil {
		t.Fatal("Reconciler should not be nil")
	}

	// Verify it can validate targets
	err = ctx.Reconciler.Validate(ctx.Targets)
	if err != nil {
		t.Errorf("Reconciler should validate targets: %v", err)
	}
}

// TestInitializeCommand_BackupOverride verifies backup flag override works
func TestInitializeCommand_BackupOverride(t *testing.T) {
	cli := &CLI{
		Globals: Globals{
			Config:   "../testdata/01-example-config.cue",
			StateDir: "",
			DryRun:   true,
		},
	}

	backup := true
	ctx, err := InitializeCommand(cli, nil, nil, &backup)
	if err != nil {
		t.Fatalf("InitializeCommand failed: %v", err)
	}

	// Check that file targets have backup enabled
	for _, target := range ctx.Targets {
		if target.GetType() == types.TYPE_FILE {
			fileTarget, ok := target.(*file.Target)
			if !ok {
				continue
			}
			if !fileTarget.GetConfig().Backup {
				t.Error("backup should be enabled for file targets")
			}
		}
	}
}
