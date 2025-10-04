package features_test

import (
	"testing"

	"github.com/thedataflows/confedit/internal/features"
	"github.com/thedataflows/confedit/internal/features/file"
	"github.com/thedataflows/confedit/internal/types"
)

func TestRegistry_Register(t *testing.T) {
	registry := features.NewRegistry()
	fileFeature := file.New()

	registry.Register(fileFeature)

	if !registry.Has(types.TYPE_FILE) {
		t.Error("file feature should be registered")
	}
}

func TestRegistry_Get(t *testing.T) {
	registry := features.NewRegistry()
	fileFeature := file.New()
	registry.Register(fileFeature)

	feature, err := registry.Get(types.TYPE_FILE)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if feature.Type() != types.TYPE_FILE {
		t.Errorf("expected type %s, got %s", types.TYPE_FILE, feature.Type())
	}
}

func TestRegistry_GetUnknown(t *testing.T) {
	registry := features.NewRegistry()

	_, err := registry.Get("unknown")
	if err == nil {
		t.Error("expected error for unknown target type")
	}
}

func TestRegistry_Executor(t *testing.T) {
	registry := features.NewRegistry()
	fileFeature := file.New()
	registry.Register(fileFeature)

	executor, err := registry.Executor(types.TYPE_FILE)
	if err != nil {
		t.Fatalf("Executor() error = %v", err)
	}

	if executor == nil {
		t.Error("executor should not be nil")
	}
}

func TestRegistry_Types(t *testing.T) {
	registry := features.NewRegistry()
	fileFeature := file.New()
	registry.Register(fileFeature)

	registeredTypes := registry.Types()
	if len(registeredTypes) != 1 {
		t.Errorf("expected 1 type, got %d", len(registeredTypes))
	}

	if registeredTypes[0] != "file" {
		t.Errorf("expected 'file', got %s", registeredTypes[0])
	}
}

func TestRegistry_MultipleFeatures(t *testing.T) {
	registry := features.NewRegistry()

	// Register file feature for this test
	fileFeature := file.New()
	registry.Register(fileFeature)

	// Verify we can register and retrieve all feature types
	expectedCount := 1 // Only file is registered in this test
	registeredTypes := registry.Types()

	if len(registeredTypes) != expectedCount {
		t.Errorf("expected %d types, got %d", expectedCount, len(registeredTypes))
	}

	// Verify the feature exists
	if !registry.Has("file") {
		t.Error("file feature should be registered")
	}
}
