package features

import (
	"fmt"

	"github.com/thedataflows/confedit/internal/engine"
)

// Registry manages all available features (file, dconf, systemd, sed)
// It provides a central point for feature discovery and executor retrieval
type Registry struct {
	features map[string]Feature
}

// NewRegistry creates a new feature registry
// All features must be registered during initialization
func NewRegistry() *Registry {
	return &Registry{
		features: make(map[string]Feature),
	}
}

// Register adds a feature to the registry
// The feature's Type() is used as the key
func (r *Registry) Register(feature Feature) {
	r.features[feature.Type()] = feature
}

// Get retrieves a feature by type
func (r *Registry) Get(targetType string) (Feature, error) {
	feature, exists := r.features[targetType]
	if !exists {
		return nil, fmt.Errorf("unknown target type: %s", targetType)
	}
	return feature, nil
}

// Executor retrieves the executor for a specific target type
// This is a convenience method commonly used by the reconciliation engine
func (r *Registry) Executor(targetType string) (engine.Executor, error) {
	feature, err := r.Get(targetType)
	if err != nil {
		return nil, err
	}
	return feature.Executor(), nil
}

// Has checks if a feature is registered for the given type
func (r *Registry) Has(targetType string) bool {
	_, exists := r.features[targetType]
	return exists
}

// Types returns all registered target types
func (r *Registry) Types() []string {
	types := make([]string, 0, len(r.features))
	for targetType := range r.features {
		types = append(types, targetType)
	}
	return types
}

// Features returns all registered features
func (r *Registry) Features() []Feature {
	features := make([]Feature, 0, len(r.features))
	for _, feature := range r.features {
		features = append(features, feature)
	}
	return features
}
