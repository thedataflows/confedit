package dconf

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/thedataflows/confedit/internal/engine"
	"github.com/thedataflows/confedit/internal/state"
	"github.com/thedataflows/confedit/internal/types"
)

// Executor implements the engine.Executor interface for dconf targets
type Executor struct{}

// NewExecutor creates a new dconf executor
func NewExecutor() engine.Executor {
	return &Executor{}
}

// Apply applies the configuration changes to dconf
func (e *Executor) Apply(target types.AnyTarget, diff *state.ConfigDiff) error {
	if diff != nil && diff.IsEmpty() {
		return nil
	}

	if target.GetType() != types.TYPE_DCONF {
		return fmt.Errorf("expected dconf target, got %s", target.GetType())
	}

	dconfTarget, ok := target.(*Target)
	if !ok {
		return fmt.Errorf("target is not a dconf target")
	}

	if dconfTarget.GetConfig() == nil {
		return fmt.Errorf("dconf target configuration is missing")
	}

	schema := dconfTarget.GetConfig().Schema
	if schema == "" {
		return fmt.Errorf("dconf schema not specified")
	}

	// Apply only the changed keys from the target settings
	settings := dconfTarget.GetConfig().Settings
	for key := range diff.Changes {
		value, exists := settings[key]
		if !exists {
			continue
		}

		cmd := exec.Command("dconf", "write",
			fmt.Sprintf("%s/%s", schema, key),
			fmt.Sprintf("'%v'", value))

		if dconfTarget.GetConfig().User != "" {
			cmd.Env = append(os.Environ(),
				fmt.Sprintf("SUDO_USER=%s", dconfTarget.GetConfig().User))
		}

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("set dconf key %s: %w", key, err)
		}
	}
	return nil
}

// Validate checks if the target configuration is valid
func (e *Executor) Validate(target types.AnyTarget) error {
	if target.GetType() != types.TYPE_DCONF {
		return fmt.Errorf("expected dconf target, got %s", target.GetType())
	}

	dconfTarget, ok := target.(*Target)
	if !ok {
		return fmt.Errorf("target is not a dconf target")
	}

	if dconfTarget.GetConfig() == nil {
		return fmt.Errorf("dconf target configuration is missing")
	}

	if dconfTarget.GetConfig().Schema == "" {
		return fmt.Errorf("dconf schema is required")
	}
	return nil
}

// GetCurrentState retrieves the current state from dconf
func (e *Executor) GetCurrentState(target types.AnyTarget) (map[string]interface{}, error) {
	if target.GetType() != types.TYPE_DCONF {
		return nil, fmt.Errorf("expected dconf target, got %s", target.GetType())
	}

	dconfTarget, ok := target.(*Target)
	if !ok {
		return nil, fmt.Errorf("target is not a dconf target")
	}

	if dconfTarget.GetConfig() == nil {
		return nil, fmt.Errorf("dconf target configuration is missing")
	}

	schema := dconfTarget.GetConfig().Schema
	if schema == "" {
		return make(map[string]interface{}), fmt.Errorf("dconf schema not specified")
	}

	// Get current dconf values for the schema
	cmd := exec.Command("dconf", "dump", schema)

	if dconfTarget.GetConfig().User != "" {
		cmd.Env = append(os.Environ(),
			fmt.Sprintf("SUDO_USER=%s", dconfTarget.GetConfig().User))
	}

	output, err := cmd.Output()
	if err != nil {
		return make(map[string]interface{}), nil // Return empty if can't read
	}

	// Simple implementation - store raw output
	result := make(map[string]interface{})
	result["_raw"] = string(output)

	return result, nil
}

// Verify that Executor implements the engine.Executor interface at compile time
var _ engine.Executor = (*Executor)(nil)
