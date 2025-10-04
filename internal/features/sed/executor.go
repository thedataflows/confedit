package sed

import (
	"fmt"
	"io"
	"os"
	"strings"

	goSed "github.com/rwtodd/Go.Sed/sed"
	"github.com/thedataflows/confedit/internal/engine"
	"github.com/thedataflows/confedit/internal/features/file"
	"github.com/thedataflows/confedit/internal/state"
	"github.com/thedataflows/confedit/internal/types"
)

// Executor implements the engine.Executor interface for sed targets
type Executor struct{}

// NewExecutor creates a new sed executor
func NewExecutor() engine.Executor {
	return &Executor{}
}

// Apply applies sed commands to the target file
func (e *Executor) Apply(target types.AnyTarget, diff *state.ConfigDiff) error {
	if diff != nil && diff.IsEmpty() {
		return nil
	}

	if target.GetType() != types.TYPE_SED {
		return fmt.Errorf("expected sed target, got %s", target.GetType())
	}

	sedTarget, ok := target.(*Target)
	if !ok {
		return fmt.Errorf("target is not a sed target")
	}

	if sedTarget.GetConfig() == nil {
		return fmt.Errorf("sed target configuration is missing")
	}

	config := sedTarget.GetConfig()
	if config.Path == "" {
		return fmt.Errorf("sed target path is required")
	}

	if len(config.Commands) == 0 {
		return fmt.Errorf("sed commands are required")
	}

	// Create backup if requested
	if config.Backup {
		if err := file.CreateBackup(config.Path); err != nil {
			return fmt.Errorf("create backup: %w", err)
		}
	}

	// Open the file for reading
	fileHandle, err := os.Open(config.Path)
	if err != nil {
		return fmt.Errorf("open file %s: %w", config.Path, err)
	}
	defer fileHandle.Close()

	// Apply sed commands using memory buffer
	script := strings.Join(config.Commands, "\n")
	sedEngine, err := goSed.New(strings.NewReader(script))
	if err != nil {
		return fmt.Errorf("create sed engine: %w", err)
	}

	// Process content in memory
	var output strings.Builder
	if _, err := io.Copy(&output, sedEngine.Wrap(fileHandle)); err != nil {
		return fmt.Errorf("run sed commands: %w", err)
	}

	// Write processed content directly to the original file
	if err := os.WriteFile(config.Path, []byte(output.String()), 0644); err != nil {
		return fmt.Errorf("write processed content: %w", err)
	}

	return nil
}

// Validate checks if the target configuration is valid
func (e *Executor) Validate(target types.AnyTarget) error {
	if target.GetType() != types.TYPE_SED {
		return fmt.Errorf("expected sed target, got %s", target.GetType())
	}

	sedTarget, ok := target.(*Target)
	if !ok {
		return fmt.Errorf("target is not a sed target")
	}

	if sedTarget.GetConfig() == nil {
		return fmt.Errorf("sed target configuration is missing")
	}

	config := sedTarget.GetConfig()
	if config.Path == "" {
		return fmt.Errorf("sed target path is required")
	}

	if len(config.Commands) == 0 {
		return fmt.Errorf("sed commands are required")
	}

	// Test if commands are valid by creating a sed engine
	script := strings.Join(config.Commands, "\n")
	_, err := goSed.New(strings.NewReader(script))
	if err != nil {
		return fmt.Errorf("invalid sed commands: %w", err)
	}

	return nil
}

// GetCurrentState retrieves the current state of the target file
func (e *Executor) GetCurrentState(target types.AnyTarget) (map[string]interface{}, error) {
	if target.GetType() != types.TYPE_SED {
		return nil, fmt.Errorf("expected sed target, got %s", target.GetType())
	}

	sedTarget, ok := target.(*Target)
	if !ok {
		return nil, fmt.Errorf("target is not a sed target")
	}

	if sedTarget.GetConfig() == nil {
		return nil, fmt.Errorf("sed target configuration is missing")
	}

	config := sedTarget.GetConfig()
	if config.Path == "" {
		return nil, fmt.Errorf("sed target path is required")
	}

	// Read current file content
	content, err := os.ReadFile(config.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]interface{}{
				"content": "",
				"exists":  false,
			}, nil
		}
		return nil, fmt.Errorf("read file %s: %w", config.Path, err)
	}

	return map[string]interface{}{
		"content": string(content),
		"exists":  true,
	}, nil
}

// Verify that Executor implements the engine.Executor interface at compile time
var _ engine.Executor = (*Executor)(nil)
