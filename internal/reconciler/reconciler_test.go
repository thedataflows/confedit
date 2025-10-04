package reconciler_test

import (
	"testing"

	"github.com/thedataflows/confedit/internal/features"
	"github.com/thedataflows/confedit/internal/features/dconf"
	"github.com/thedataflows/confedit/internal/features/file"
	"github.com/thedataflows/confedit/internal/features/sed"
	"github.com/thedataflows/confedit/internal/features/systemd"
	"github.com/thedataflows/confedit/internal/reconciler"
	"github.com/thedataflows/confedit/internal/state"
	"github.com/thedataflows/confedit/internal/types"
)

func TestReconciliationEngine_Creation(t *testing.T) {
	// Create feature registry
	registry := features.NewRegistry()
	registry.Register(file.New())
	registry.Register(dconf.New())
	registry.Register(systemd.New())
	registry.Register(sed.New())

	// Create state manager
	stateManager := state.NewManager("")

	// Create reconciler
	r := reconciler.NewReconciliationEngine(registry, stateManager, false)

	if r == nil {
		t.Fatal("reconciler should not be nil")
	}
}

func TestReconciliationEngine_Validate(t *testing.T) {
	// Create feature registry
	registry := features.NewRegistry()
	registry.Register(file.New())

	// Create state manager
	stateManager := state.NewManager("")

	// Create reconciler
	r := reconciler.NewReconciliationEngine(registry, stateManager, false)

	// Create a valid file target
	target := file.NewTarget("test", "/tmp/test.ini", "ini")

	// Validate should succeed
	err := r.Validate([]types.AnyTarget{target})
	if err != nil {
		t.Fatalf("validation should succeed: %v", err)
	}
}

func TestReconciliationEngine_ValidateInvalidTarget(t *testing.T) {
	// Create feature registry
	registry := features.NewRegistry()
	registry.Register(file.New())

	// Create state manager
	stateManager := state.NewManager("")

	// Create reconciler
	r := reconciler.NewReconciliationEngine(registry, stateManager, false)

	// Create an invalid file target (unsupported format)
	target := &file.Target{
		Name: "test",
		Type: types.TYPE_FILE,
		Config: &file.Config{
			Path:   "/tmp/test.txt",
			Format: "unsupported",
		},
	}

	// Validate should fail
	err := r.Validate([]types.AnyTarget{target})
	if err == nil {
		t.Fatal("validation should fail for invalid target")
	}
}
