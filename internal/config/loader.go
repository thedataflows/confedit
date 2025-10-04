package config

// CUE configuration loader

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"github.com/thedataflows/confedit/internal/features/dconf"
	"github.com/thedataflows/confedit/internal/features/file"
	"github.com/thedataflows/confedit/internal/features/sed"
	"github.com/thedataflows/confedit/internal/features/systemd"
	"github.com/thedataflows/confedit/internal/types"
	"github.com/thedataflows/confedit/internal/utils"
	log "github.com/thedataflows/go-lib-log"
)

type CueConfigLoader struct {
	configPath string
	ctx        *cue.Context
	validator  *SchemaValidator
}

func NewCueConfigLoader(configPath string, schemaFilePath ...string) *CueConfigLoader {
	validator, err := NewSchemaValidator(schemaFilePath...)
	if err != nil {
		log.Logger().Fatal().Err(err).Msg("Failed to initialize schema validator")
	}

	return &CueConfigLoader{
		configPath: configPath,
		ctx:        cuecontext.New(),
		validator:  validator,
	}
}

func (ccl *CueConfigLoader) LoadConfiguration() (*types.SystemConfig, error) {
	filePaths, err := ccl.collectCueFiles()
	if err != nil {
		return nil, fmt.Errorf("collect CUE files: %w", err)
	}

	mergedConfig := &types.SystemConfig{
		Targets:   []types.AnyTarget{},
		Variables: make(map[string]interface{}),
	}
	targetsByName := make(map[string]types.AnyTarget)

	// Process each file
	for _, filePath := range filePaths {
		workingDir := filepath.Dir(filePath)
		fileName := filepath.Base(filePath)

		// Load CUE file
		value, err := ccl.loadAndBuildCUE(workingDir, fileName)
		if err != nil {
			return nil, fmt.Errorf("load CUE file '%s': %w", filePath, err)
		}

		// Decode CUE value
		fileConfig, err := ccl.decodeCUEValue(value)
		if err != nil {
			return nil, fmt.Errorf("decode CUE config '%s': %w", filePath, err)
		}

		// Merge targets
		for _, target := range fileConfig.Targets {
			if existingTarget, exists := targetsByName[target.GetName()]; exists {
				// Merge with existing target
				if err := ccl.mergeTargets(existingTarget, target); err != nil {
					return nil, fmt.Errorf("merge target '%s' in '%s': %w", target.GetName(), filePath, err)
				}
			} else {
				// Add new target
				targetsByName[target.GetName()] = target
				mergedConfig.Targets = append(mergedConfig.Targets, target)
			}
		}

		// Merge variables
		maps.Copy(mergedConfig.Variables, fileConfig.Variables)

		// Merge hooks (last one wins)
		if fileConfig.Hooks != nil {
			mergedConfig.Hooks = fileConfig.Hooks
		}
	}

	// Validate the configuration
	if err := ccl.ValidateConfiguration(); err != nil {
		return nil, fmt.Errorf("config validation: %w", err)
	}

	return mergedConfig, nil
}

// collectCueFiles gathers all .cue files from the config path (file or directory) in lexicographical order
func (ccl *CueConfigLoader) collectCueFiles() ([]string, error) {
	stat, err := os.Stat(ccl.configPath)
	if err != nil {
		return nil, err
	}

	var filePaths []string

	if stat.IsDir() {
		// Directory: collect all .cue files recursively
		err := filepath.Walk(ccl.configPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(info.Name(), ".cue") {
				filePaths = append(filePaths, path)
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("walk directory '%s': %w", ccl.configPath, err)
		}
	} else {
		// Single file
		if !strings.HasSuffix(ccl.configPath, ".cue") {
			return nil, fmt.Errorf("file must have .cue extension: %s", ccl.configPath)
		}
		filePaths = append(filePaths, ccl.configPath)
	}

	// Sort lexicographically
	sort.Strings(filePaths)
	return filePaths, nil
}

// ValidateConfiguration validates configuration files without loading them
func (ccl *CueConfigLoader) ValidateConfiguration() error {
	if ccl.validator == nil {
		return fmt.Errorf("schema validator not available")
	}

	filePaths, err := ccl.collectCueFiles()
	if err != nil {
		return fmt.Errorf("failed to collect CUE files: %w", err)
	}

	for _, filePath := range filePaths {
		workingDir := filepath.Dir(filePath)
		fileName := filepath.Base(filePath)

		buildInstances := load.Instances([]string{fileName}, &load.Config{
			Dir: workingDir,
		})

		values, err := ccl.ctx.BuildInstances(buildInstances)
		if err != nil {
			return fmt.Errorf("failed to build CUE instance for '%s': %w", filePath, err)
		}

		if len(values) != 1 {
			return fmt.Errorf("expected exactly one CUE value for '%s'", filePath)
		}

		value := values[0]
		if err := value.Err(); err != nil {
			return fmt.Errorf("CUE value error in '%s': %w", filePath, err)
		}

		valueStr, err := value.MarshalJSON()
		if err != nil {
			return fmt.Errorf("failed to marshal value for validation in '%s': %w", filePath, err)
		}

		if err := ccl.validator.ValidateRawConfig(valueStr); err != nil {
			return fmt.Errorf("validation failed for '%s': %w", filePath, err)
		}
	}

	return nil
}

// mergeTargets deeply merges a new target into an existing target with the same name
func (ccl *CueConfigLoader) mergeTargets(existing types.AnyTarget, newTarget types.AnyTarget) error {
	// Merge metadata first
	if err := utils.DeepMerge(existing.GetMetadata(), newTarget.GetMetadata()); err != nil {
		return fmt.Errorf("merge metadata: %w", err)
	}

	// Ensure both targets are the same type
	if existing.GetType() != newTarget.GetType() {
		return fmt.Errorf("cannot merge targets of different types: %s vs %s", existing.GetType(), newTarget.GetType())
	}

	// Use a simple switch to merge the config - much cleaner than complex generics
	switch existing.GetType() {
	case types.TYPE_FILE:
		existingFile := existing.(*file.Target)
		newFile := newTarget.(*file.Target)
		return ccl.mergeFileTargets(existingFile.Config, newFile.Config)
	case types.TYPE_DCONF:
		existingDconf := existing.(*dconf.Target)
		newDconf := newTarget.(*dconf.Target)
		return ccl.mergeDconfTargets(existingDconf.Config, newDconf.Config)
	case types.TYPE_SYSTEMD:
		existingSystemd := existing.(*systemd.Target)
		newSystemd := newTarget.(*systemd.Target)
		return ccl.mergeSystemdTargets(existingSystemd.Config, newSystemd.Config)
	default:
		return fmt.Errorf("unsupported target type for merging: %s", existing.GetType())
	}
}

// mergeFileTargets merges file target configurations using map merging
func (ccl *CueConfigLoader) mergeFileTargets(existing, newTarget *file.Config) error {
	// Merge all the map-based content
	if err := utils.DeepMerge(existing.Content, newTarget.Content); err != nil {
		return fmt.Errorf("merge content: %w", err)
	}
	if err := utils.DeepMerge(existing.Options, newTarget.Options); err != nil {
		return fmt.Errorf("merge options: %w", err)
	}

	// Update scalar fields (new values override existing ones if non-empty)
	if newTarget.Path != "" {
		existing.Path = newTarget.Path
	}
	if newTarget.Format != "" {
		existing.Format = newTarget.Format
	}
	if newTarget.Owner != "" {
		existing.Owner = newTarget.Owner
	}
	if newTarget.Group != "" {
		existing.Group = newTarget.Group
	}
	if newTarget.Mode != "" {
		existing.Mode = newTarget.Mode
	}
	if newTarget.Backup {
		existing.Backup = true
	}

	return nil
}

// mergeDconfTargets merges dconf target configurations using map merging
func (ccl *CueConfigLoader) mergeDconfTargets(existing, newTarget *dconf.Config) error {
	if err := utils.DeepMerge(existing.Settings, newTarget.Settings); err != nil {
		return fmt.Errorf("merge settings: %w", err)
	}

	if newTarget.Schema != "" {
		existing.Schema = newTarget.Schema
	}
	if newTarget.User != "" {
		existing.User = newTarget.User
	}

	return nil
}

// mergeSystemdTargets merges systemd target configurations using map merging
func (ccl *CueConfigLoader) mergeSystemdTargets(existing, newTarget *systemd.Config) error {
	if err := utils.DeepMerge(existing.Properties, newTarget.Properties); err != nil {
		return fmt.Errorf("merge properties: %w", err)
	}

	if newTarget.Unit != "" {
		existing.Unit = newTarget.Unit
	}
	if newTarget.Section != "" {
		existing.Section = newTarget.Section
	}
	if newTarget.Reload {
		existing.Reload = true
	}

	return nil
}

// loadAndBuildCUE loads and builds a CUE instance from file
func (ccl *CueConfigLoader) loadAndBuildCUE(workingDir, targetFile string) (cue.Value, error) {
	instances := load.Instances([]string{targetFile}, &load.Config{
		Dir: workingDir,
	})

	if len(instances) == 0 {
		return cue.Value{}, fmt.Errorf("no CUE instances found")
	}

	values, err := ccl.ctx.BuildInstances(instances)
	if err != nil {
		return cue.Value{}, fmt.Errorf("build CUE instance: %w", err)
	}

	if len(values) == 0 {
		return cue.Value{}, fmt.Errorf("no CUE values produced")
	}

	value := values[0]
	if err := value.Err(); err != nil {
		return cue.Value{}, fmt.Errorf("CUE value error: %w", err)
	}

	return value, nil
}

// CommonTargetFields represents the common fields for all target types
type CommonTargetFields struct {
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// decodeCUEValue decodes a CUE value into SystemConfig
func (ccl *CueConfigLoader) decodeCUEValue(value cue.Value) (*types.SystemConfig, error) {
	var fileConfig types.SystemConfig
	fileConfig.Targets = []types.AnyTarget{}
	fileConfig.Variables = make(map[string]interface{})

	// Decode variables
	variablesValue := value.LookupPath(cue.ParsePath("variables"))
	if variablesValue.Exists() {
		if err := variablesValue.Decode(&fileConfig.Variables); err != nil {
			return nil, fmt.Errorf("decode variables: %w", err)
		}
	}

	// Decode hooks
	hooksValue := value.LookupPath(cue.ParsePath("hooks"))
	if hooksValue.Exists() {
		if err := hooksValue.Decode(&fileConfig.Hooks); err != nil {
			return nil, fmt.Errorf("decode hooks: %w", err)
		}
	}

	// Decode targets
	targetsValue := value.LookupPath(cue.ParsePath("targets"))
	if targetsValue.Exists() {
		iter, err := targetsValue.List()
		if err != nil {
			return nil, fmt.Errorf("iterate targets: %w", err)
		}

		for iter.Next() {
			target, err := ccl.decodeTarget(iter.Value())
			if err != nil {
				return nil, err
			}
			fileConfig.Targets = append(fileConfig.Targets, target)
		}
	}

	return &fileConfig, nil
}

// decodeTarget decodes a single target from CUE value
func (ccl *CueConfigLoader) decodeTarget(targetValue cue.Value) (types.AnyTarget, error) {
	// Decode common fields
	var commonFields CommonTargetFields

	if err := targetValue.Decode(&commonFields); err != nil {
		return nil, fmt.Errorf("decode target common fields: %w", err)
	}

	// Initialize metadata if nil
	if commonFields.Metadata == nil {
		commonFields.Metadata = make(map[string]interface{})
	}

	// Get config value
	configValue := targetValue.LookupPath(cue.ParsePath("config"))
	if !configValue.Exists() {
		return nil, fmt.Errorf("%s target missing 'config' configuration", commonFields.Type)
	}

	// Create target using the type registry
	return ccl.createTarget(commonFields, configValue)
}

// createTarget creates a target directly without registry - simplified implementation
func (ccl *CueConfigLoader) createTarget(commonFields CommonTargetFields, configValue cue.Value) (types.AnyTarget, error) {
	// Create and decode target config based on type
	switch commonFields.Type {
	case types.TYPE_FILE:
		config := &file.Config{
			Content: make(map[string]interface{}),
			Options: make(map[string]interface{}),
		}
		if err := configValue.Decode(config); err != nil {
			return nil, fmt.Errorf("decode %s target config: %w", commonFields.Type, err)
		}
		return &file.Target{
			Name:     commonFields.Name,
			Type:     commonFields.Type,
			Metadata: commonFields.Metadata,
			Config:   config,
		}, nil

	case types.TYPE_DCONF:
		config := &dconf.Config{
			Settings: make(map[string]interface{}),
		}
		if err := configValue.Decode(config); err != nil {
			return nil, fmt.Errorf("decode %s target config: %w", commonFields.Type, err)
		}
		return &dconf.Target{
			Name:     commonFields.Name,
			Type:     commonFields.Type,
			Metadata: commonFields.Metadata,
			Config:   config,
		}, nil

	case types.TYPE_SYSTEMD:
		config := &systemd.Config{
			Properties: make(map[string]interface{}),
		}
		if err := configValue.Decode(config); err != nil {
			return nil, fmt.Errorf("decode %s target config: %w", commonFields.Type, err)
		}
		return &systemd.Target{
			Name:     commonFields.Name,
			Type:     commonFields.Type,
			Metadata: commonFields.Metadata,
			Config:   config,
		}, nil

	case types.TYPE_SED:
		config := &sed.Config{
			Commands: make([]string, 0),
		}
		if err := configValue.Decode(config); err != nil {
			return nil, fmt.Errorf("decode %s target config: %w", commonFields.Type, err)
		}
		return &sed.Target{
			Name:     commonFields.Name,
			Type:     commonFields.Type,
			Metadata: commonFields.Metadata,
			Config:   config,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported target type: %s", commonFields.Type)
	}
}
