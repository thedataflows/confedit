package iniparser

import (
	"io"
)

// INIWrapper implements FormatParser for INI files with structure preservation
type INIWrapper struct {
	parser *RelaxedINIParser
	lines  []INILine // Preserved lines from last parse for structure preservation
}

// NewINIWrapper creates a new INI wrapper that implements FormatParser
func NewINIWrapper() *INIWrapper {
	return &INIWrapper{
		parser: NewRelaxedINIParser(),
	}
}

// Configure implements ConfigurableParser interface to accept options.
// Supported options:
//   - use_spacing (bool): Controls delimiter formatting for new keys.
//     When true (default), new keys are written as "key = value".
//     When false, new keys are written as "key=value".
//     Existing keys always preserve their original formatting.
//   - comment_chars (string): Characters to recognize as comment prefixes.
//     Default is "#;" (both # and ; are recognized).
//     Each character in the string will be treated as a valid comment prefix.
//   - delimiter (string): Key-value delimiter character.
//     Default is "=". Only the first character is used if multiple are provided.
func (w *INIWrapper) Configure(options map[string]interface{}) error {
	if options == nil {
		return nil
	}

	// Handle use_spacing option
	if useSpacing, ok := options["use_spacing"]; ok {
		if boolVal, ok := useSpacing.(bool); ok {
			w.parser.useSpacing = boolVal
		}
	}

	// Handle comment_chars option
	if commentChars, ok := options["comment_chars"]; ok {
		if strVal, ok := commentChars.(string); ok && len(strVal) > 0 {
			w.parser.commentChars = []byte(strVal)
		}
	}

	// Handle delimiter option
	if delimiter, ok := options["delimiter"]; ok {
		if strVal, ok := delimiter.(string); ok && len(strVal) > 0 {
			w.parser.delimiter = strVal[0]
		}
	}

	return nil
}

// Unmarshal parses INI data and returns a nested map structure
// Structure: map[section]map[key]value
// Special cases:
// - Root keys (no section) are stored under "" (empty string) section
// - Commented lines with key=value structure are NOT parsed into the map (per design)
// - Only active (uncommented) key-value pairs are included
// - To parse commented lines, use ParseLine() to extract key/value manually
func (w *INIWrapper) Unmarshal(data []byte) (map[string]interface{}, error) {
	lines, err := w.parser.Parse(data)
	if err != nil {
		return nil, err
	}
	w.lines = lines

	result := make(map[string]interface{})

	for _, line := range lines {
		// Skip non-key lines (empty lines, pure comments, section headers)
		if line.Key == "" {
			continue
		}

		// Skip commented lines - they are not parsed into structured data
		// Caller should use ParseLine() if they need to extract key/value from comments
		if line.CommentPrefix != "" {
			continue
		}

		// Ensure section exists
		section := w.ensureSection(result, line.Section)

		// Store active key-value pairs only
		section[line.Key] = line.Value
	}

	return result, nil
}

// Marshal writes the map structure back to INI format
// If lines were previously parsed, it updates them in place to preserve formatting
// Otherwise, it builds new lines from scratch
func (w *INIWrapper) Marshal(data map[string]interface{}, writer io.Writer) error {
	var lines []INILine

	if len(w.lines) > 0 {
		// Update existing lines to preserve structure and order
		lines = w.updateLines(data)
	} else {
		// Build from scratch - iterate data in natural map order
		lines = w.buildLines(data)
	}

	return w.parser.Serialize(lines, writer)
}

// ensureSection gets or creates a section map in the result
func (w *INIWrapper) ensureSection(result map[string]interface{}, section string) map[string]interface{} {
	if sectionMap, ok := result[section].(map[string]interface{}); ok {
		return sectionMap
	}

	sectionMap := make(map[string]interface{})
	result[section] = sectionMap
	return sectionMap
}

// updateLines updates existing lines based on new data
func (w *INIWrapper) updateLines(data map[string]interface{}) []INILine {
	lines := make([]INILine, 0, len(w.lines))
	processed := make(map[string]bool)
	currentSection := ""

	// First pass: update existing lines
	for _, line := range w.lines {
		if line.IsSection {
			currentSection = line.Section
			lines = append(lines, line)
			continue
		}

		// Preserve non-key lines (empty lines, pure comments without keys)
		if line.Key == "" {
			lines = append(lines, line)
			continue
		}

		// Preserve commented lines as-is (they're not in the data map by design)
		if line.CommentPrefix != "" {
			lines = append(lines, line)
			continue
		}

		// Mark active key as processed
		key := makeKey(currentSection, line.Key)
		processed[key] = true

		// Find value in data
		value := findValue(data, currentSection, line.Key)
		if value == nil {
			// Key not in data - skip it (deletion)
			continue
		}

		// Update line with new value
		updatedLine := updateLineValue(line, value)
		lines = append(lines, updatedLine)
	}

	// Second pass: add new keys not in original
	newKeys := w.collectNewKeys(data, processed)
	if len(newKeys) > 0 {
		lines = w.insertNewKeys(lines, newKeys)
	}

	return lines
}

// buildLines creates lines from scratch when no previous structure exists
// Iterates data in natural map order (no sorting to preserve simplicity)
func (w *INIWrapper) buildLines(data map[string]interface{}) []INILine {
	var lines []INILine

	// Process root section first if it exists
	if rootData, ok := data[""].(map[string]interface{}); ok {
		for key, value := range rootData {
			line := createLine("", key, value)
			// Skip deleted keys (empty line marker)
			if line.Key == "" {
				continue
			}
			lines = append(lines, line)
		}
	}

	// Process all other sections
	for sectionName, sectionData := range data {
		if sectionName == "" {
			continue // Already processed
		}

		sectionMap, ok := sectionData.(map[string]interface{})
		if !ok || len(sectionMap) == 0 {
			continue
		}

		// Add section header
		lines = append(lines, INILine{
			Section:   sectionName,
			IsSection: true,
		})

		// Add keys in natural map order
		for key, value := range sectionMap {
			line := createLine(sectionName, key, value)
			// Skip deleted keys (empty line marker)
			if line.Key == "" {
				continue
			}
			lines = append(lines, line)
		}
	}

	return lines
}

// makeKey creates a unique key for tracking (section::key)
func makeKey(section, key string) string {
	return section + "::" + key
}

// findValue finds a value in the nested data structure
func findValue(data map[string]interface{}, section, key string) interface{} {
	sectionData, ok := data[section].(map[string]interface{})
	if !ok {
		return nil
	}
	return sectionData[key]
}

// updateLineValue updates a line with new value
func updateLineValue(line INILine, value interface{}) INILine {
	// Handle deletion marker
	if valueMap, ok := value.(map[string]interface{}); ok {
		if deleted, exists := valueMap["deleted"]; exists && deleted == true {
			// Marked for deletion - return empty (will be filtered out)
			return INILine{}
		}
	}

	// Regular value update
	if strValue, ok := value.(string); ok {
		line.Value = strValue
		line.CommentPrefix = "" // Ensure it's uncommented
	}

	return line
}

// createLine creates a new INILine from section, key, value
func createLine(section, key string, value interface{}) INILine {
	// Handle deletion marker
	if valueMap, ok := value.(map[string]interface{}); ok {
		if deleted, exists := valueMap["deleted"]; exists && deleted == true {
			// Marked for deletion - return empty line (will be filtered out)
			return INILine{}
		}
	}

	line := INILine{
		Section: section,
		Key:     key,
	}

	// Regular value
	if strValue, ok := value.(string); ok {
		line.Value = strValue
	}

	return line
}

// collectNewKeys finds keys in data that weren't in original lines
func (w *INIWrapper) collectNewKeys(data map[string]interface{}, processed map[string]bool) map[string][]keyValue {
	newKeys := make(map[string][]keyValue)

	for sectionName, sectionData := range data {
		sectionMap, ok := sectionData.(map[string]interface{})
		if !ok {
			continue
		}

		for key, value := range sectionMap {
			uniqueKey := makeKey(sectionName, key)
			if !processed[uniqueKey] {
				newKeys[sectionName] = append(newKeys[sectionName], keyValue{
					key:   key,
					value: value,
				})
			}
		}
	}

	return newKeys
}

// keyValue represents a key-value pair
type keyValue struct {
	key   string
	value interface{}
}

// insertNewKeys inserts new keys into appropriate sections
func (w *INIWrapper) insertNewKeys(lines []INILine, newKeys map[string][]keyValue) []INILine {
	result := make([]INILine, 0, len(lines)+len(newKeys))
	currentSection := ""

	for i, line := range lines {
		result = append(result, line)

		if line.IsSection {
			currentSection = line.Section
		}

		// Check if we should insert new keys after this line
		if keys, exists := newKeys[currentSection]; exists {
			// Find last key position in section
			if isLastKeyInSection(lines, i) {
				for _, kv := range keys {
					newLine := createLine(currentSection, kv.key, kv.value)
					result = append(result, newLine)
				}
				delete(newKeys, currentSection) // Mark as processed
			}
		}
	}

	// Add remaining keys for sections that weren't found
	for sectionName, keys := range newKeys {
		if sectionName != "" {
			// Add section header
			result = append(result, INILine{
				Section:   sectionName,
				IsSection: true,
			})
		}
		for _, kv := range keys {
			result = append(result, createLine(sectionName, kv.key, kv.value))
		}
	}

	return result
}

// isLastKeyInSection checks if this is the last key in the current section
// by scanning forward from the given index to see if there are any more active keys
// before hitting a section boundary or end of file
func isLastKeyInSection(lines []INILine, index int) bool {
	for i := index + 1; i < len(lines); i++ {
		line := lines[i]
		if line.IsSection {
			// Hit next section
			return true
		}
		if line.Key != "" && line.CommentPrefix == "" {
			// Found another active key in same section
			return false
		}
	}
	// End of file
	return true
}

// Parse is a convenience method that wraps Unmarshal
func (w *INIWrapper) Parse(data []byte) (map[string]interface{}, error) {
	return w.Unmarshal(data)
}

// Serialize is a convenience method that wraps Marshal
func (w *INIWrapper) Serialize(data map[string]interface{}, writer io.Writer) error {
	return w.Marshal(data, writer)
}

// ParseLine exposes the underlying parser's line parsing for external use
// This allows callers to parse commented lines to extract section/key/value
func (w *INIWrapper) ParseLine(lineBytes []byte, section string) INILine {
	return w.parser.parseLine(lineBytes, section)
}
