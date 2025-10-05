package iniparser

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testDataDir = "testdata"

// loadTestFile loads a test file from the testdata directory
func loadTestFile(t *testing.T, filename string) string {
	t.Helper()
	path := filepath.Join(testDataDir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("load test file %s: %v", filename, err)
	}
	return string(data)
}

// getAllTestFiles returns all .ini files in the testdata directory
func getAllTestFiles(t *testing.T) []string {
	t.Helper()
	testdataDir := filepath.Join(testDataDir)
	entries, err := os.ReadDir(testdataDir)
	if err != nil {
		t.Fatalf("read testdata directory: %v", err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && (strings.HasSuffix(entry.Name(), ".ini") || strings.HasSuffix(entry.Name(), ".conf")) {
			// Skip expected output files for modification tests
			if !strings.Contains(entry.Name(), "_expected") {
				files = append(files, entry.Name())
			}
		}
	}
	return files
}

// =============================================================================
// Structure Preservation Tests
// =============================================================================

func TestINIParser_PreservesExactStructure(t *testing.T) {
	files := getAllTestFiles(t)
	parser := NewRelaxedINIParser()

	for _, filename := range files {
		// temporary, fopr debugging
		// if filename != "pacman-light.conf" {
		// 	continue // Skip this file for now
		// }
		t.Run(filename, func(tt *testing.T) {
			input := loadTestFile(tt, filename)

			// Parse the input
			parsed, err := parser.Parse([]byte(input))
			if err != nil {
				tt.Fatalf("Parse failed for %s: %v", filename, err)
			}

			// Serialize back without any modifications
			var buf bytes.Buffer
			err = parser.Serialize(parsed, &buf)
			if err != nil {
				tt.Fatalf("Serialize failed for %s: %v", filename, err)
			}

			result := strings.TrimSpace(buf.String())
			expected := strings.TrimSpace(input)

			if result != expected {
				tt.Errorf("Structure not preserved for %s!\nExpected:\n%s\n\nGot:\n%s\n", filename, expected, result)
			}
		})
	}
}

// =============================================================================
// Value Modification Tests
// =============================================================================

func TestINIParser_PreservesStructureAfterValueModification(t *testing.T) {
	files := getAllTestFiles(t)
	parser := NewRelaxedINIParser()

	for _, filename := range files {
		t.Run(filename, func(tt *testing.T) {
			input := loadTestFile(tt, filename)

			// Parse the input
			parsed, err := parser.Parse([]byte(input))
			if err != nil {
				tt.Fatalf("Parse failed for %s: %v", filename, err)
			}

			// Modify existing values where possible
			modified := false
			for i := range parsed {
				line := &parsed[i]
				if line.Key != "" && line.Value != "" && line.CommentPrefix == "" {
					line.Value = "modified_" + line.Value
					modified = true
					break // Only modify one value per file to test structure preservation
				}
			}

			if !modified {
				tt.Skip("No modifiable key-value pairs found in " + filename)
			}

			// Serialize back
			var buf bytes.Buffer
			err = parser.Serialize(parsed, &buf)
			if err != nil {
				tt.Fatalf("Serialize failed for %s after modification: %v", filename, err)
			}

			result := buf.String()

			// Verify structure elements are preserved
			originalLines := strings.Split(input, "\n")
			resultLines := strings.Split(result, "\n")

			// Count comment lines, empty lines, and sections to ensure they're preserved
			countComments := func(lines []string) int {
				count := 0
				for _, line := range lines {
					if strings.HasPrefix(strings.TrimSpace(line), "#") || strings.HasPrefix(strings.TrimSpace(line), ";") {
						count++
					}
				}
				return count
			}

			countEmptyLines := func(lines []string) int {
				count := 0
				for _, line := range lines {
					if strings.TrimSpace(line) == "" {
						count++
					}
				}
				return count
			}

			countSections := func(lines []string) int {
				count := 0
				for _, line := range lines {
					trimmed := strings.TrimSpace(line)
					if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
						count++
					}
				}
				return count
			}

			// Verify structural elements are preserved
			if countComments(originalLines) != countComments(resultLines) {
				tt.Errorf("Comment count changed for %s after modification", filename)
			}
			if countEmptyLines(originalLines) != countEmptyLines(resultLines) {
				tt.Errorf("Empty line count changed for %s after modification", filename)
			}
			if countSections(originalLines) != countSections(resultLines) {
				tt.Errorf("Section count changed for %s after modification", filename)
			}

			// Verify the modification was applied
			if !strings.Contains(result, "modified_") {
				tt.Errorf("Value modification not applied for %s", filename)
			}
		})
	}
}

// =============================================================================
// Key Addition Tests
// =============================================================================

func TestINIParser_PreservesStructureAfterKeyAddition(t *testing.T) {
	files := getAllTestFiles(t)
	parser := NewRelaxedINIParser()

	for _, filename := range files {
		t.Run(filename, func(tt *testing.T) {
			input := loadTestFile(tt, filename)

			// Parse the input
			parsed, err := parser.Parse([]byte(input))
			if err != nil {
				tt.Fatalf("Parse failed for %s: %v", filename, err)
			}

			// Find the first section and add a new key
			sectionFound := ""
			for _, line := range parsed {
				if line.IsSection {
					sectionFound = line.Section
					break
				}
			}

			if sectionFound == "" {
				tt.Skip("No sections found to add key to in " + filename)
			}

			// Add a new key using the parser's AddKey method
			parsed = parser.AddKey(parsed, sectionFound, "test_new_key", "test_new_value")

			// Serialize back
			var buf bytes.Buffer
			err = parser.Serialize(parsed, &buf)
			if err != nil {
				tt.Fatalf("Serialize failed for %s after key addition: %v", filename, err)
			}

			result := buf.String()

			// Verify the new key was added
			if !strings.Contains(result, "test_new_key") {
				tt.Errorf("New key not added for %s", filename)
			}
			if !strings.Contains(result, "test_new_value") {
				tt.Errorf("New value not added for %s", filename)
			}

			// Verify original content is still present
			originalLines := strings.Split(input, "\n")
			for _, line := range originalLines {
				trimmed := strings.TrimSpace(line)
				if trimmed != "" && !strings.HasPrefix(trimmed, "#") && !strings.HasPrefix(trimmed, ";") {
					// For non-comment, non-empty lines, check if the essence is preserved
					if strings.Contains(trimmed, "=") || strings.Contains(trimmed, ":") {
						// It's a key-value line, check if key is still present
						parts := strings.FieldsFunc(trimmed, func(c rune) bool { return c == '=' || c == ':' })
						if len(parts) > 0 {
							key := strings.TrimSpace(parts[0])
							if key != "" && !strings.Contains(result, key) {
								tt.Errorf("Original key '%s' missing after key addition in %s", key, filename)
							}
						}
					} else if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
						// It's a section header
						if !strings.Contains(result, trimmed) {
							tt.Errorf("Original section '%s' missing after key addition in %s", trimmed, filename)
						}
					}
				}
			}
		})
	}
}

// =============================================================================
// Utility Function Tests
// =============================================================================

func TestINIParser_UpdateValue(t *testing.T) {
	parser := NewRelaxedINIParser()
	input := `[section1]
key1=value1
key2=value2

[section2]
key3=value3`

	parsed, err := parser.Parse([]byte(input))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Update a value
	parsed = parser.UpdateValue(parsed, "section1", "key1", "new_value1")

	// Serialize back
	var buf bytes.Buffer
	err = parser.Serialize(parsed, &buf)
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	result := buf.String()

	// Verify the value was updated
	if !strings.Contains(result, "key1=new_value1") && !strings.Contains(result, "key1 = new_value1") {
		t.Errorf("Value not updated. Result:\n%s", result)
	}

	// Verify other values remain unchanged
	if !strings.Contains(result, "key2=value2") && !strings.Contains(result, "key2 = value2") {
		t.Errorf("Other values changed unexpectedly. Result:\n%s", result)
	}
}

func TestINIParser_GetValue(t *testing.T) {
	parser := NewRelaxedINIParser()
	input := `[section1]
key1=value1
key2=value2

[section2]
key3=value3`

	parsed, err := parser.Parse([]byte(input))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Test getting existing values
	value, exists := GetValue(parsed, "section1", "key1")
	if !exists {
		t.Errorf("Expected key1 to exist in section1")
	}
	if value != "value1" {
		t.Errorf("Expected value1, got %s", value)
	}

	// Test getting non-existent value
	_, exists = GetValue(parsed, "section1", "nonexistent")
	if exists {
		t.Errorf("Expected nonexistent key to not exist")
	}

	// Test getting value from different section
	value, exists = GetValue(parsed, "section2", "key3")
	if !exists {
		t.Errorf("Expected key3 to exist in section2")
	}
	if value != "value3" {
		t.Errorf("Expected value3, got %s", value)
	}
}

// =============================================================================
// Edge Cases Tests
// =============================================================================

func TestINIParser_EmptyLines(t *testing.T) {
	parser := NewRelaxedINIParser()
	input := `[section1]
key1=value1

key2=value2


[section2]

key3=value3`

	parsed, err := parser.Parse([]byte(input))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Serialize back
	var buf bytes.Buffer
	err = parser.Serialize(parsed, &buf)
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	result := strings.TrimSpace(buf.String())
	expected := strings.TrimSpace(input)

	if result != expected {
		t.Errorf("Empty lines not preserved!\nExpected:\n%s\n\nGot:\n%s\n", expected, result)
	}
}

func TestINIParser_BareKeys(t *testing.T) {
	parser := NewRelaxedINIParser()
	input := `[section1]
key1=value1
bare_key
key2=value2`

	parsed, err := parser.Parse([]byte(input))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Serialize back
	var buf bytes.Buffer
	err = parser.Serialize(parsed, &buf)
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	result := strings.TrimSpace(buf.String())
	expected := strings.TrimSpace(input)

	if result != expected {
		t.Errorf("Bare keys not preserved!\nExpected:\n%s\n\nGot:\n%s\n", expected, result)
	}
}

func TestINIParser_InlineComments(t *testing.T) {
	parser := NewRelaxedINIParser()
	input := `[section1]
key1=value1  # This is a comment
key2=value2  ; This is another comment
key3=value3`

	parsed, err := parser.Parse([]byte(input))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Serialize back
	var buf bytes.Buffer
	err = parser.Serialize(parsed, &buf)
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	result := strings.TrimSpace(buf.String())
	expected := strings.TrimSpace(input)

	if result != expected {
		t.Errorf("Inline comments not preserved!\nExpected:\n%s\n\nGot:\n%s\n", expected, result)
	}
}

func TestINIParser_CommentedKeys(t *testing.T) {
	parser := NewRelaxedINIParser()
	input := `[section1]
key1=value1
#key2=value2
;key3=value3
key4=value4`

	parsed, err := parser.Parse([]byte(input))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Serialize back
	var buf bytes.Buffer
	err = parser.Serialize(parsed, &buf)
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	result := strings.TrimSpace(buf.String())
	expected := strings.TrimSpace(input)

	if result != expected {
		t.Errorf("Commented keys not preserved!\nExpected:\n%s\n\nGot:\n%s\n", expected, result)
	}
}

func TestParseLine(t1 *testing.T) {
	parser := NewRelaxedINIParser()

	tests := []struct {
		name      string
		input     string
		wantEmpty bool
		wantSec   bool
		comment   string
		indent    string
		section   string
		key       string
		delim     string
		value     string
		suffix    string
	}{
		{
			name:      "empty line",
			input:     "",
			wantEmpty: true,
		},
		{
			name:      "spaces only",
			input:     "   \t  ",
			wantEmpty: true,
			indent:    "   \t  ",
		},
		{
			name:    "comment with hash",
			input:   "# This is a comment",
			comment: "# ",
			key:     "This is a comment",
		},
		{
			name:    "comment with semicolon",
			input:   "; This is a comment",
			comment: "; ",
			key:     "This is a comment",
		},
		{
			name:    "comment with indent",
			input:   "  ; Indented comment",
			comment: "; ",
			indent:  "  ",
			key:     "Indented comment",
		},
		{
			name:    "section",
			input:   "[database]",
			wantSec: true,
			section: "database",
		},
		{
			name:    "section with indent",
			input:   "  [app]",
			wantSec: true,
			indent:  "  ",
			section: "app",
		},
		{
			name:    "section with suffix",
			input:   "[options] # inline comment",
			wantSec: true,
			section: "options",
			suffix:  " # inline comment",
		},
		{
			name:  "key=value",
			input: "host=localhost",
			key:   "host",
			delim: "=",
			value: "localhost",
		},
		{
			name:  "key = value with spaces",
			input: "port = 5432",
			key:   "port",
			delim: " = ",
			value: "5432",
		},
		{
			name:   "key with leading/trailing spaces",
			input:  "  timeout  =  30  ",
			indent: "  ",
			key:    "timeout",
			delim:  "  =  ",
			value:  "30  ",
		},
		{
			name:  "bare key",
			input: "enable_feature",
			key:   "enable_feature",
		},
		{
			name:    "commented key=value",
			input:   "; host=localhost",
			comment: "; ",
			key:     "host=localhost", // Commented lines are NOT parsed - entire rest is stored as Key
		},
		{
			name:    "commented section",
			input:   "# [disabled]",
			comment: "# ",
			key:     "[disabled]", // Commented lines are NOT parsed - entire rest is stored as Key
		},
		{
			name:  "key with dots",
			input: "server.host = example.com",
			key:   "server.host",
			delim: " = ",
			value: "example.com",
		},
		{
			name:  "key with dash",
			input: "retry-count=3",
			key:   "retry-count",
			delim: "=",
			value: "3",
		},
		{
			name:  "value with equals sign",
			input: "password=p@ss=word!123",
			key:   "password",
			delim: "=",
			value: "p@ss=word!123",
		},
	}

	for _, tt := range tests {
		t1.Run(tt.name, func(t2 *testing.T) {
			result := parser.parseLine([]byte(tt.input), "")

			if result.IsEmpty != tt.wantEmpty {
				t2.Errorf("IsEmpty = %v, want %v", result.IsEmpty, tt.wantEmpty)
			}
			if result.IsSection != tt.wantSec {
				t2.Errorf("IsSection = %v, want %v", result.IsSection, tt.wantSec)
			}
			if result.CommentPrefix != tt.comment {
				t2.Errorf("CommentPrefix = %q, want %q", result.CommentPrefix, tt.comment)
			}
			if result.Indent != tt.indent {
				t2.Errorf("Indent = %q, want %q", result.Indent, tt.indent)
			}
			if result.Section != tt.section {
				t2.Errorf("Section = %q, want %q", result.Section, tt.section)
			}
			if result.Key != tt.key {
				t2.Errorf("Key = %q, want %q", result.Key, tt.key)
			}
			if result.Delimiter != tt.delim {
				t2.Errorf("Delimiter = %q, want %q", result.Delimiter, tt.delim)
			}
			if result.Value != tt.value {
				t2.Errorf("Value = %q, want %q", result.Value, tt.value)
			}
			if result.Suffix != tt.suffix {
				t2.Errorf("Suffix = %q, want %q", result.Suffix, tt.suffix)
			}
		})
	}
}
