package file

import (
	"bytes"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/thedataflows/confedit/internal/engine"
	"github.com/thedataflows/confedit/internal/features/file/formats"
	"github.com/thedataflows/confedit/internal/state"
	"github.com/thedataflows/confedit/internal/types"
	"github.com/thedataflows/confedit/internal/utils"
	log "github.com/thedataflows/go-lib-log"
)

// Executor implements the engine.Executor interface for file targets
type Executor struct {
	registry *formats.Registry
}

// NewExecutor creates a new file executor with the given format registry
func NewExecutor(registry *formats.Registry) engine.Executor {
	return &Executor{
		registry: registry,
	}
}

// Apply applies the configuration changes to the target file
func (e *Executor) Apply(target types.AnyTarget, diff *state.ConfigDiff) error {
	if target.GetType() != types.TYPE_FILE {
		return fmt.Errorf("expected file target, got %s", target.GetType())
	}

	fileTarget, ok := target.(*Target)
	if !ok {
		return fmt.Errorf("target is not a file target")
	}

	// If there are no changes, skip the operation (idempotency check)
	if diff != nil && diff.IsEmpty() {
		log.Debugf("file-executor", "No changes needed for file: %s", fileTarget.GetConfig().Path)
		return nil
	}

	if fileTarget.Config == nil {
		return fmt.Errorf("file target configuration is missing")
	}

	log.Debugf("file-executor", "Applying changes to file: %s", fileTarget.GetConfig().Path)

	format := fileTarget.GetConfig().Format
	if format == "" {
		format = "ini" // default format
	}

	parser, err := e.registry.Get(format)
	if err != nil {
		return fmt.Errorf("get parser: %w", err)
	}

	// Configure parser with format-specific options if it supports configuration
	if configurableParser, ok := parser.(formats.ConfigurableParser); ok {
		if err := configurableParser.Configure(fileTarget.GetConfig().Options); err != nil {
			return fmt.Errorf("configure parser: %w", err)
		}
	}

	// Create backup if requested
	if fileTarget.GetConfig().Backup {
		if err := CreateBackup(fileTarget.GetConfig().Path); err != nil {
			return fmt.Errorf("create backup: %w", err)
		}
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(fileTarget.GetConfig().Path), 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	// Get current file state to preserve unmanaged keys
	currentState, err := e.GetCurrentState(target)
	if err != nil {
		return fmt.Errorf("get current state: %w", err)
	}

	// Merge desired content into current state to preserve all unmanaged keys
	if err := utils.DeepMerge(currentState, fileTarget.GetConfig().Content); err != nil {
		return fmt.Errorf("merge content: %w", err)
	}

	// Marshal and write the patched state
	var buf bytes.Buffer
	if err := parser.Marshal(currentState, &buf); err != nil {
		return fmt.Errorf("marshal content: %w", err)
	}

	if err := os.WriteFile(fileTarget.GetConfig().Path, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	// Set ownership and permissions
	if err := e.setFileOwnership(fileTarget.GetConfig()); err != nil {
		return fmt.Errorf("set ownership: %w", err)
	}

	if err := e.setFilePermissions(fileTarget.GetConfig()); err != nil {
		return fmt.Errorf("set permissions: %w", err)
	}

	log.Debugf("file-executor", "Successfully updated file: %s", fileTarget.GetConfig().Path)
	return nil
}

// Validate checks if the target configuration is valid
func (e *Executor) Validate(target types.AnyTarget) error {
	if target.GetType() != types.TYPE_FILE {
		return fmt.Errorf("expected file target, got %s", target.GetType())
	}

	fileTarget, ok := target.(*Target)
	if !ok {
		return fmt.Errorf("target is not a file target")
	}

	if fileTarget.GetConfig() == nil {
		return fmt.Errorf("file target configuration is missing")
	}

	// Check if format is supported
	format := fileTarget.GetConfig().Format
	if format == "" {
		format = "ini" // default format
	}
	if !e.registry.Has(format) {
		return fmt.Errorf("unsupported file format: %s", format)
	}

	// Check if path is valid
	if fileTarget.GetConfig().Path == "" {
		return fmt.Errorf("target path is empty")
	}

	return nil
}

// GetCurrentState retrieves the current state of the target file
func (e *Executor) GetCurrentState(target types.AnyTarget) (map[string]interface{}, error) {
	if target.GetType() != types.TYPE_FILE {
		return nil, fmt.Errorf("expected file target, got %s", target.GetType())
	}

	fileTarget, ok := target.(*Target)
	if !ok {
		return nil, fmt.Errorf("target is not a file target")
	}

	if fileTarget.GetConfig() == nil {
		return nil, fmt.Errorf("file target configuration is missing")
	}

	if _, err := os.Stat(fileTarget.GetConfig().Path); os.IsNotExist(err) {
		return make(map[string]interface{}), nil
	}

	format := fileTarget.GetConfig().Format
	if format == "" {
		return nil, fmt.Errorf("file format is not specified")
	}

	parser, err := e.registry.Get(format)
	if err != nil {
		return nil, fmt.Errorf("get parser: %w", err)
	}

	// Configure parser with format-specific options if it supports configuration
	if configurableParser, ok := parser.(formats.ConfigurableParser); ok {
		if err := configurableParser.Configure(fileTarget.GetConfig().Options); err != nil {
			return nil, fmt.Errorf("configure parser: %w", err)
		}
	}

	data, err := os.ReadFile(fileTarget.GetConfig().Path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	return parser.Unmarshal(data)
}

// setFileOwnership sets the owner and group of the file
func (e *Executor) setFileOwnership(target *Config) error {
	owner := target.Owner
	group := target.Group

	if owner == "" && group == "" {
		return nil
	}

	var uid, gid int

	if owner != "" {
		u, err := user.Lookup(owner)
		if err != nil {
			return fmt.Errorf("lookup user %s: %w", owner, err)
		}
		uid, _ = strconv.Atoi(u.Uid)
	} else {
		uid = -1
	}

	if group != "" {
		g, err := user.LookupGroup(group)
		if err != nil {
			return fmt.Errorf("lookup group %s: %w", group, err)
		}
		gid, _ = strconv.Atoi(g.Gid)
	} else {
		gid = -1
	}

	return syscall.Chown(target.Path, uid, gid)
}

// setFilePermissions sets the file permissions
func (e *Executor) setFilePermissions(target *Config) error {
	modeValue := target.Mode
	if modeValue == "" {
		return nil
	}

	mode, err := strconv.ParseUint(modeValue, 8, 32)
	if err != nil {
		return fmt.Errorf("invalid file mode %s: %w", modeValue, err)
	}

	return os.Chmod(target.Path, os.FileMode(mode))
}

// Verify that Executor implements the engine.Executor interface at compile time
var _ engine.Executor = (*Executor)(nil)
