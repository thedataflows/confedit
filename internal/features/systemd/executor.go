package systemd

import (
	"fmt"
	"os/exec"

	"github.com/thedataflows/confedit/internal/engine"
	"github.com/thedataflows/confedit/internal/state"
	"github.com/thedataflows/confedit/internal/types"
)

// Executor implements the engine.Executor interface for systemd targets
type Executor struct{}

// NewExecutor creates a new systemd executor
func NewExecutor() engine.Executor {
	return &Executor{}
}

// Apply applies the configuration changes to systemd
func (e *Executor) Apply(target types.AnyTarget, diff *state.ConfigDiff) error {
	if diff != nil && diff.IsEmpty() {
		return nil
	}

	if target.GetType() != types.TYPE_SYSTEMD {
		return fmt.Errorf("expected systemd target, got %s", target.GetType())
	}

	systemdTarget, ok := target.(*Target)
	if !ok {
		return fmt.Errorf("target is not a systemd target")
	}

	if systemdTarget.GetConfig() == nil {
		return fmt.Errorf("systemd target configuration is missing")
	}

	unitFile := systemdTarget.GetConfig().Unit
	if unitFile == "" {
		return fmt.Errorf("systemd unit file not specified")
	}

	// Update unit file with changes from target properties
	if len(diff.Changes) > 0 {
		err := e.updateUnitFile(unitFile, systemdTarget.GetConfig().Properties)
		if err != nil {
			return err
		}

		// Reload systemd daemon
		cmd := exec.Command("systemctl", "daemon-reload")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("reload systemd: %w", err)
		}
	}

	// Handle service state changes
	if systemdTarget.GetConfig().Reload {
		cmd := exec.Command("systemctl", "reload-or-restart", unitFile)
		return cmd.Run()
	}

	return nil
}

// Validate checks if the target configuration is valid
func (e *Executor) Validate(target types.AnyTarget) error {
	if target.GetType() != types.TYPE_SYSTEMD {
		return fmt.Errorf("expected systemd target, got %s", target.GetType())
	}

	systemdTarget, ok := target.(*Target)
	if !ok {
		return fmt.Errorf("target is not a systemd target")
	}

	if systemdTarget.GetConfig() == nil {
		return fmt.Errorf("systemd target configuration is missing")
	}

	if systemdTarget.GetConfig().Unit == "" {
		return fmt.Errorf("systemd unit is required")
	}
	return nil
}

// GetCurrentState retrieves the current state from systemd
func (e *Executor) GetCurrentState(target types.AnyTarget) (map[string]interface{}, error) {
	if target.GetType() != types.TYPE_SYSTEMD {
		return nil, fmt.Errorf("expected systemd target, got %s", target.GetType())
	}

	systemdTarget, ok := target.(*Target)
	if !ok {
		return nil, fmt.Errorf("target is not a systemd target")
	}

	if systemdTarget.GetConfig() == nil {
		return nil, fmt.Errorf("systemd target configuration is missing")
	}

	unitFile := systemdTarget.GetConfig().Unit
	if unitFile == "" {
		return make(map[string]interface{}), fmt.Errorf("systemd unit file not specified")
	}

	// Get unit file status
	cmd := exec.Command("systemctl", "show", unitFile)
	output, err := cmd.Output()
	if err != nil {
		return make(map[string]interface{}), nil
	}

	result := make(map[string]interface{})
	result["_status"] = string(output)

	return result, nil
}

// updateUnitFile updates the systemd unit file
func (e *Executor) updateUnitFile(unitFile string, changes map[string]interface{}) error {
	// Simplified implementation - in real scenario, parse and update unit file
	return fmt.Errorf("unit file update not implemented")
}

// Verify that Executor implements the engine.Executor interface at compile time
var _ engine.Executor = (*Executor)(nil)
