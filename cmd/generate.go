package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/format"
	"github.com/alecthomas/kong"
	"github.com/thedataflows/confedit/internal/engine"
	"github.com/thedataflows/confedit/internal/features"
	"github.com/thedataflows/confedit/internal/features/dconf"
	"github.com/thedataflows/confedit/internal/features/file"
	"github.com/thedataflows/confedit/internal/features/sed"
	"github.com/thedataflows/confedit/internal/features/systemd"
	"github.com/thedataflows/confedit/internal/types"
	log "github.com/thedataflows/go-lib-log"
)

// File format constants
const (
	FORMAT_INI  = "ini"
	FORMAT_YAML = "yaml"
	FORMAT_TOML = "toml"
	FORMAT_JSON = "json"
	FORMAT_XML  = "xml"
)

// GenerateCmd generates a .cue data file from the diff between two executor targets
type GenerateCmd struct {
	Source      []string `arg:"" help:"Source and target executor specifications (format: type:path)"`
	Type        string   `short:"t" required:"" help:"Type of the target executor" enum:"file,dconf,systemd,sed"`
	Name        string   `short:"n" help:"Name for the generated target"`
	Output      string   `short:"o" help:"Output file path for the generated .cue data"`
	Identifiers string   `help:"Target identifiers in flattened format (e.g., options.use_spacing=true,backup=false,metadata.custom=value)"`
	FileFormat  string   `help:"File format for file targets (e.g., ini, yaml, toml, json, xml). Work only with file targets. Overrides auto-detected format."`
	registry    *features.Registry
}

// NewGenerateCmd creates a new generate command

func (c *GenerateCmd) Run(ctx *kong.Context, cli *CLI) error {
	if len(c.Source) != 2 {
		return fmt.Errorf("exactly two arguments required: source and target executor specifications")
	}

	log.Infof(PKG_CMD, "Generating diff between '%s' and '%s'", c.Source[0], c.Source[1])

	if c.Name == "" {
		c.Name = normalizeNameFromSource(c.Source[1])
	}

	// Create source executor and get state
	sourceExecutor, err := c.createExecutor(c.Type)
	if err != nil {
		return fmt.Errorf("create source executor: %w", err)
	}

	sourceTarget := c.createTarget(c.Source[0])
	sourceState, err := sourceExecutor.CurrentState(sourceTarget)
	if err != nil {
		return fmt.Errorf("get source state: %w", err)
	}

	// Create destination executor and get state (might be empty if file doesn't exist)
	destExecutor, err := c.createExecutor(c.Type)
	if err != nil {
		return fmt.Errorf("create destination executor: %w", err)
	}

	destTarget := c.createTarget(c.Source[1])
	destState, err := destExecutor.CurrentState(destTarget)
	if err != nil {
		// If destination doesn't exist, use empty state
		destState = make(map[string]interface{})
	}

	// Compute simple diff
	diff := c.computeSimpleDiff(destState, sourceState)
	if len(diff) == 0 {
		log.Warnf(PKG_CMD, "No differences found between states")
		return nil
	}

	// Generate CUE data
	cueData := c.generateSimpleCUE(c.Source[1], diff)

	// Determine output path
	outputPath := c.Output
	if outputPath == "" {
		if cli.Config == "" {
			outputPath = c.Name + ".cue"
		} else {
			if isDirectory(cli.Config) {
				outputPath = filepath.Join(cli.Config, c.Name+".cue")
			} else {
				outputPath = cli.Config
			}
		}
	}

	// Write output
	if err := c.writeOutput(outputPath, cueData); err != nil {
		return fmt.Errorf("write output to %s: %w", outputPath, err)
	}

	log.Infof(PKG_CMD, "Successfully generated: %s", outputPath)
	return nil
}

// createTarget creates a simple target for the given path
func (c *GenerateCmd) createTarget(path string) types.AnyTarget {
	switch c.Type {
	case types.TYPE_FILE:
		// Auto-detect format based on file extension
		fileFormat := c.detectFormat(path)
		return file.NewTarget("target", path, fileFormat)
	case types.TYPE_DCONF:
		return dconf.NewTarget("target", path)
	case types.TYPE_SYSTEMD:
		return systemd.NewTarget("target", path, "")
	case types.TYPE_SED:
		return sed.NewTarget("target", path, []string{})
	default:
		return nil
	}
}

// computeSimpleDiff computes a simple diff between source and destination states
// Returns only the keys that differ, with their destination (new) values
func (c *GenerateCmd) computeSimpleDiff(source, dest map[string]interface{}) map[string]interface{} {
	diff := make(map[string]interface{})

	// Check all keys in dest - include if value differs from source or is new
	for key, destValue := range dest {
		sourceValue, existsInSource := source[key]
		if !existsInSource || !c.valuesEqual(sourceValue, destValue) {
			diff[key] = c.computeValueDiff(sourceValue, destValue, existsInSource)
		}
	}

	return diff
}

// computeValueDiff handles nested map comparisons for accurate diffs
func (c *GenerateCmd) computeValueDiff(sourceVal, destVal interface{}, sourceExists bool) interface{} {
	// If source doesn't exist, return the entire dest value (new key)
	if !sourceExists {
		return destVal
	}

	// If both are maps, recursively compute diff for nested structure
	sourceMap, sourceIsMap := sourceVal.(map[string]interface{})
	destMap, destIsMap := destVal.(map[string]interface{})

	if sourceIsMap && destIsMap {
		nestedDiff := make(map[string]interface{})

		// Check all keys in dest map
		for key, dVal := range destMap {
			sVal, existsInSource := sourceMap[key]
			if !existsInSource || !c.valuesEqual(sVal, dVal) {
				nestedDiff[key] = c.computeValueDiff(sVal, dVal, existsInSource)
			}
		}

		// Only return the nested diff if there are actual differences
		if len(nestedDiff) > 0 {
			return nestedDiff
		}
		return nil
	}

	// For non-map values, return the dest value if they differ
	return destVal
}

// valuesEqual compares two values for equality
func (c *GenerateCmd) valuesEqual(a, b interface{}) bool {
	// Simple equality check - can be enhanced if needed
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

// generateSimpleCUE generates a simple CUE data using CUE encoder
func (c *GenerateCmd) generateSimpleCUE(targetPath string, diff map[string]interface{}) string {
	// Create the data structure
	output := map[string]interface{}{
		"targets": []map[string]interface{}{
			{
				"name":   c.Name,
				"type":   c.Type,
				"config": c.buildConfig(targetPath, diff),
			},
		},
	}

	// Use CUE context to encode
	ctx := cuecontext.New()
	val := ctx.Encode(output)

	// Format as CUE
	formatted, err := format.Node(val.Syntax(), format.Simplify())
	if err != nil {
		// Fallback to basic formatting if CUE formatting fails
		return fmt.Sprintf("package config\n\n%v", output)
	}

	return fmt.Sprintf("package config\n\n%s", formatted)
}

// buildConfig builds the config section for a target
func (c *GenerateCmd) buildConfig(targetPath string, diff map[string]interface{}) map[string]interface{} {
	config := map[string]interface{}{
		"path": targetPath,
	}

	// Handle different target types
	switch c.Type {
	case types.TYPE_SED:
		// For sed targets, convert diff to commands
		if commands, ok := diff["commands"]; ok {
			config["commands"] = commands
		} else {
			// If no commands found, provide empty array
			config["commands"] = []string{}
		}
	case types.TYPE_FILE:
		config["content"] = diff
		// Add format based on file extension
		fileFormat := c.detectFormat(targetPath)
		if fileFormat != "" {
			config["format"] = fileFormat
		}
	default:
		config["content"] = diff
	}

	return config
}

// detectFormat detects file format based on extension
func (c *GenerateCmd) detectFormat(filePath string) string {
	if c.FileFormat != "" {
		return c.FileFormat
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".ini", ".conf":
		return FORMAT_INI
	case ".yaml", ".yml":
		return FORMAT_YAML
	case ".toml":
		return FORMAT_TOML
	case ".json":
		return FORMAT_JSON
	case ".xml":
		return FORMAT_XML
	default:
		return ""
	}
}

// initializeRegistry creates and registers all available features
func (c *GenerateCmd) initializeRegistry() {
	if c.registry == nil {
		c.registry = features.NewRegistry()
		c.registry.Register(file.New())
		c.registry.Register(dconf.New())
		c.registry.Register(sed.New())
		c.registry.Register(systemd.New())
	}
}

// createExecutor creates an executor instance for the given type
func (c *GenerateCmd) createExecutor(execType string) (engine.Executor, error) {
	c.initializeRegistry()
	return c.registry.Executor(execType)
}

// writeOutput writes the generated CUE data to the output file
func (c *GenerateCmd) writeOutput(outputPath, content string) error {
	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	// Write the file
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

// isDirectory checks if a path is a directory
func isDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// normalizeNameFromSource normalizes a source path to be used as a name
// by converting all special characters to dashes
func normalizeNameFromSource(source string) string {
	if source == "" {
		return ""
	}

	var result strings.Builder
	result.Grow(len(source)) // Pre-allocate capacity

	for _, r := range source {
		// Keep alphanumeric characters, convert everything else to dash
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			result.WriteRune(r)
		} else {
			result.WriteByte('-')
		}
	}

	// Clean up multiple consecutive dashes and trim edge dashes
	normalized := result.String()

	// Replace multiple consecutive dashes with single dash
	for strings.Contains(normalized, "--") {
		normalized = strings.ReplaceAll(normalized, "--", "-")
	}

	// Trim leading and trailing dashes
	normalized = strings.Trim(normalized, "-")

	return normalized
}
