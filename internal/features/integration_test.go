package features_test

import (
	"testing"

	"github.com/thedataflows/confedit/internal/features"
	"github.com/thedataflows/confedit/internal/features/dconf"
	"github.com/thedataflows/confedit/internal/features/file"
	"github.com/thedataflows/confedit/internal/features/sed"
	"github.com/thedataflows/confedit/internal/features/systemd"
)

func TestAllFeatures_Registration(t *testing.T) {
	registry := features.NewRegistry()

	// Register all features
	registry.Register(file.New())
	registry.Register(dconf.New())
	registry.Register(systemd.New())
	registry.Register(sed.New())

	// Verify all features are registered
	expectedTypes := map[string]bool{
		"file":    true,
		"dconf":   true,
		"systemd": true,
		"sed":     true,
	}

	registeredTypes := registry.Types()
	if len(registeredTypes) != len(expectedTypes) {
		t.Errorf("expected %d types, got %d", len(expectedTypes), len(registeredTypes))
	}

	for _, typeName := range registeredTypes {
		if !expectedTypes[typeName] {
			t.Errorf("unexpected type registered: %s", typeName)
		}
		delete(expectedTypes, typeName)
	}

	if len(expectedTypes) > 0 {
		for typeName := range expectedTypes {
			t.Errorf("expected type not registered: %s", typeName)
		}
	}
}

func TestAllFeatures_GetExecutor(t *testing.T) {
	registry := features.NewRegistry()

	// Register all features
	registry.Register(file.New())
	registry.Register(dconf.New())
	registry.Register(systemd.New())
	registry.Register(sed.New())

	// Test getting executors for each feature
	featureTypes := []string{"file", "dconf", "systemd", "sed"}

	for _, featureType := range featureTypes {
		t.Run(featureType, func(tb *testing.T) {
			executor, err := registry.Executor(featureType)
			if err != nil {
				tb.Fatalf("get executor for %s: %v", featureType, err)
			}
			if executor == nil {
				tb.Fatalf("executor for %s is nil", featureType)
			}
		})
	}
}

func TestAllFeatures_GetFeature(t *testing.T) {
	registry := features.NewRegistry()

	// Register all features
	registry.Register(file.New())
	registry.Register(dconf.New())
	registry.Register(systemd.New())
	registry.Register(sed.New())

	// Test getting features
	featureTypes := []string{"file", "dconf", "systemd", "sed"}

	for _, featureType := range featureTypes {
		t.Run(featureType, func(tb *testing.T) {
			feature, err := registry.Get(featureType)
			if err != nil {
				tb.Fatalf("get feature %s: %v", featureType, err)
			}
			if feature == nil {
				tb.Fatalf("feature %s is nil", featureType)
			}
			if feature.Type() != featureType {
				tb.Errorf("feature type mismatch: expected %s, got %s", featureType, feature.Type())
			}
		})
	}
}
