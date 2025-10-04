package state

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/thedataflows/confedit/internal/utils"
)

// ConfigDiff represents the difference between desired and current state
type ConfigDiff struct {
	Target   string                 `json:"target"`
	Changes  map[string]interface{} `json:"changes"`
	Added    map[string]interface{} `json:"added"`
	Removed  []string               `json:"removed"`
	Modified map[string]DiffValue   `json:"modified"`
}

// DiffValue represents a before/after value pair
type DiffValue struct {
	Old interface{} `json:"old"`
	New interface{} `json:"new"`
}

// ComputeDiff compares current and desired states
func ComputeDiff(current, desired map[string]interface{}) *ConfigDiff {
	diff := &ConfigDiff{
		Changes:  make(map[string]interface{}),
		Added:    make(map[string]interface{}),
		Removed:  []string{},
		Modified: make(map[string]DiffValue),
	}

	// Find added and modified keys
	for key, newValue := range desired {
		if oldValue, exists := current[key]; exists {
			if !reflect.DeepEqual(oldValue, newValue) {
				diff.Changes[key] = newValue
				diff.Modified[key] = DiffValue{Old: oldValue, New: newValue}
			}
		} else {
			diff.Changes[key] = newValue
			diff.Added[key] = newValue
		}
	}

	// Find removed keys
	for key, oldValue := range current {
		if _, exists := desired[key]; !exists {
			diff.Removed = append(diff.Removed, key)
			// For generate command, we need removed keys in Changes as well
			// This ensures that when desired is empty, we get all current state
			diff.Changes[key] = oldValue
		}
	}

	return diff
}

// IsEmpty returns true if the diff contains no changes
func (d *ConfigDiff) IsEmpty() bool {
	return len(d.Changes) == 0 && len(d.Removed) == 0
}

// ComputeChecksum computes a SHA256 checksum of the content
func ComputeChecksum(content map[string]interface{}) (string, error) {
	data, err := json.Marshal(content)
	if err != nil {
		return "", fmt.Errorf("marshal content for checksum: %w", err)
	}

	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash), nil
}

// FormatDiff returns a formatted string representation of the diff with color support
func (d *ConfigDiff) FormatDiff(colorSupport *utils.ColorSupport) string {
	if d.IsEmpty() {
		return ""
	}

	var parts []string

	// Format added keys
	if len(d.Added) > 0 {
		parts = append(parts, colorSupport.Bold("  Add:"))
		for key, value := range d.Added {
			formattedValue := formatValue(value, "    ")
			line := fmt.Sprintf("    + %s = %s", key, formattedValue)
			parts = append(parts, colorSupport.Green(line))
		}
	}

	// Format modified keys
	if len(d.Modified) > 0 {
		parts = append(parts, colorSupport.Bold("  Change:"))
		for key, diffValue := range d.Modified {
			oldValue := formatValue(diffValue.Old, "    ")
			newValue := formatValue(diffValue.New, "    ")
			line := fmt.Sprintf("    ~ %s = %s → %s", key, colorSupport.Red(oldValue), colorSupport.Green(newValue))
			parts = append(parts, colorSupport.Yellow(line))
		}
	}

	// Format removed keys
	if len(d.Removed) > 0 {
		parts = append(parts, colorSupport.Bold("  Remove:"))
		for _, key := range d.Removed {
			line := fmt.Sprintf("    - %s", key)
			parts = append(parts, colorSupport.Red(line))
		}
	}

	return strings.Join(parts, "\n")
}

// FormatPlain returns a plain text representation of the diff (no colors)
func (d *ConfigDiff) FormatPlain() string {
	// Use no-op color support to reuse FormatDiff logic
	return d.FormatDiff(&utils.ColorSupport{})
}

// FlattenForDiff flattens nested maps using dot notation for better diff display
func FlattenForDiff(data map[string]interface{}, prefix string) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		if nestedMap, ok := value.(map[string]interface{}); ok {
			// Check if this is a special INI structure (commented, deleted, etc.)
			if _, hasDeleted := nestedMap["deleted"]; hasDeleted {
				// Don't flatten deleted structures - treat them as leaf values
				result[fullKey] = value
				continue
			}
			if _, hasCommented := nestedMap["commented"]; hasCommented {
				// Don't flatten commented structures - treat them as leaf values
				result[fullKey] = value
				continue
			}
			if _, hasValue := nestedMap["value"]; hasValue {
				// Don't flatten structures with "value" field - treat them as leaf values
				result[fullKey] = value
				continue
			}

			// Recursively flatten nested maps
			flattened := FlattenForDiff(nestedMap, fullKey)
			for k, v := range flattened {
				result[k] = v
			}
		} else {
			result[fullKey] = value
		}
	}

	return result
}

// ComputeFlatDiff computes diff using flattened structures for better display
func ComputeFlatDiff(current, desired map[string]interface{}) *ConfigDiff {
	// Flatten both structures
	flatCurrent := FlattenForDiff(current, "")
	flatDesired := FlattenForDiff(desired, "")

	// Remove deleted keys from desired state - they should only show as removals
	for key, value := range flatDesired {
		if iniValue, ok := value.(map[string]interface{}); ok {
			if deleted, hasDeleted := iniValue["deleted"]; hasDeleted {
				if deletedBool, ok := deleted.(bool); ok && deletedBool {
					delete(flatDesired, key)
				}
			}
		}
	}

	return ComputeDiff(flatCurrent, flatDesired)
}

// formatValue formats a value for display
func formatValue(value interface{}, indent string) string {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf("%q", v)
	case nil:
		return "null"
	case []interface{}:
		return formatSlice(v, indent)
	case []string:
		return formatStringSlice(v, indent)
	default:
		jsonValue, _ := json.Marshal(value)
		return string(jsonValue)
	}
}

// formatSlice formats []interface{} slices
func formatSlice(items []interface{}, indent string) string {
	if len(items) == 0 {
		return "[]"
	}
	var parts []string
	for _, item := range items {
		parts = append(parts, fmt.Sprintf("%s  - %s", indent, formatValue(item, indent+"  ")))
	}
	return "[\n" + strings.Join(parts, "\n") + "\n" + indent + "]"
}

// formatStringSlice formats []string slices
func formatStringSlice(items []string, indent string) string {
	if len(items) == 0 {
		return "[]"
	}
	var parts []string
	for _, item := range items {
		parts = append(parts, fmt.Sprintf("%s  - %q", indent, item))
	}
	return "[\n" + strings.Join(parts, "\n") + "\n" + indent + "]"
}
