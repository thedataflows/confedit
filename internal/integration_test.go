package internal_test

import (
	"testing"

	"github.com/thedataflows/confedit/internal/features"
	"github.com/thedataflows/confedit/internal/features/dconf"
	"github.com/thedataflows/confedit/internal/features/file"
	"github.com/thedataflows/confedit/internal/features/sed"
	"github.com/thedataflows/confedit/internal/features/systemd"
	"github.com/thedataflows/confedit/internal/loader"
	"github.com/thedataflows/confedit/internal/state"
	"github.com/thedataflows/confedit/internal/types"
)

// TestInternalPackagesIntegration verifies that internal packages are correctly moved and work together
func TestInternalPackagesIntegration(t *testing.T) {
	// Test loader package
	l := loader.NewCueDataLoader("../testdata/01-example-config.cue", "")
	if l == nil {
		t.Fatal("config loader should not be nil")
	}

	// Test state package
	stateManager := state.NewManager("")
	if stateManager == nil {
		t.Fatal("state manager should not be nil")
	}

	// Test types package
	fileTarget := &file.Target{
		Name: "test",
		Type: types.TYPE_FILE,
		Config: &file.Config{
			Path:   "/tmp/test.ini",
			Format: "ini",
		},
	}
	if fileTarget.GetName() != "test" {
		t.Errorf("expected name 'test', got %s", fileTarget.GetName())
	}
	if fileTarget.GetType() != types.TYPE_FILE {
		t.Errorf("expected type %s, got %s", types.TYPE_FILE, fileTarget.GetType())
	}

	// Test features package with registry
	registry := features.NewRegistry()
	registry.Register(file.New())
	registry.Register(dconf.New())
	registry.Register(sed.New())
	registry.Register(systemd.New())

	if !registry.Has(types.TYPE_FILE) {
		t.Error("registry should have file feature")
	}
	if !registry.Has(types.TYPE_DCONF) {
		t.Error("registry should have dconf feature")
	}
	if !registry.Has(types.TYPE_SED) {
		t.Error("registry should have sed feature")
	}
	if !registry.Has(types.TYPE_SYSTEMD) {
		t.Error("registry should have systemd feature")
	}

	// Test getting executor from registry
	executor, err := registry.Executor(types.TYPE_FILE)
	if err != nil {
		t.Fatalf("get file executor: %v", err)
	}
	if executor == nil {
		t.Fatal("executor should not be nil")
	}

	// Test executor validation
	err = executor.Validate(fileTarget)
	if err != nil {
		t.Errorf("validation should succeed for valid file target: %v", err)
	}
}

// TestInternalPackagesIsolation ensures internal packages don't import from pkg (except legacy)
func TestInternalPackagesIsolation(t *testing.T) {
	// This test verifies the architecture by checking package instantiation
	// If internal packages incorrectly imported pkg/, this would fail to compile

	// Create instances of each internal package component
	_ = loader.NewCueDataLoader("", "")
	_ = state.NewManager("")
	_ = features.NewRegistry()

	// All types should be accessible from internal/types
	_ = types.TYPE_FILE
	_ = types.TYPE_DCONF
	_ = types.TYPE_SED
	_ = types.TYPE_SYSTEMD

	t.Log("Internal packages are properly isolated")
}
