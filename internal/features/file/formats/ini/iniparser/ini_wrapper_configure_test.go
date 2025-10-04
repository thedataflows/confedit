package iniparser

import (
	"bytes"
	"strings"
	"testing"
)

// TestINIWrapper_Configure_UseSpacing tests the Configure method with use_spacing option
func TestINIWrapper_Configure_UseSpacing(t *testing.T) {
	tests := []struct {
		name        string
		useSpacing  bool
		initialData string
		section     string
		key         string
		value       string
		expected    string
	}{
		{
			name:       "use_spacing true - adds spaces around delimiter",
			useSpacing: true,
			initialData: `[section]
existing = value
`,
			section:  "section",
			key:      "newkey",
			value:    "newvalue",
			expected: "newkey = newvalue",
		},
		{
			name:       "use_spacing false - no spaces around delimiter",
			useSpacing: false,
			initialData: `[section]
existing=value
`,
			section:  "section",
			key:      "newkey",
			value:    "newvalue",
			expected: "newkey=newvalue",
		},
		{
			name:       "use_spacing true - root section with spaces",
			useSpacing: true,
			initialData: `existing = value
`,
			section:  "",
			key:      "newkey",
			value:    "newvalue",
			expected: "newkey = newvalue",
		},
		{
			name:       "use_spacing false - root section without spaces",
			useSpacing: false,
			initialData: `existing=value
`,
			section:  "",
			key:      "newkey",
			value:    "newvalue",
			expected: "newkey=newvalue",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(subT *testing.T) {
			wrapper := NewINIWrapper()

			// Configure with use_spacing option
			err := wrapper.Configure(map[string]interface{}{
				"use_spacing": tt.useSpacing,
			})
			if err != nil {
				subT.Fatalf("Configure failed: %v", err)
			}

			// Parse initial data
			data, err := wrapper.Parse([]byte(tt.initialData))
			if err != nil {
				subT.Fatalf("Parse failed: %v", err)
			}

			// Add new key
			section := data[tt.section].(map[string]interface{})
			section[tt.key] = tt.value

			// Serialize
			var buf bytes.Buffer
			err = wrapper.Serialize(data, &buf)
			if err != nil {
				subT.Fatalf("Serialize failed: %v", err)
			}

			result := buf.String()

			// Check if the new key is formatted correctly
			if !strings.Contains(result, tt.expected) {
				subT.Errorf("Expected to find '%s' in output, got:\n%s", tt.expected, result)
			}
		})
	}
}

// TestINIWrapper_Configure_NilOptions tests Configure with nil options
func TestINIWrapper_Configure_NilOptions(t *testing.T) {
	wrapper := NewINIWrapper()

	err := wrapper.Configure(nil)
	if err != nil {
		t.Errorf("Configure with nil options should not fail, got: %v", err)
	}
}

// TestINIWrapper_Configure_MissingOption tests Configure without use_spacing option
func TestINIWrapper_Configure_MissingOption(t *testing.T) {
	wrapper := NewINIWrapper()

	err := wrapper.Configure(map[string]interface{}{
		"other_option": "value",
	})
	if err != nil {
		t.Errorf("Configure without use_spacing should not fail, got: %v", err)
	}

	// Should use default value (true)
	if !wrapper.parser.useSpacing {
		t.Errorf("Expected default useSpacing to be true")
	}
}

// TestINIWrapper_Configure_InvalidType tests Configure with wrong type for use_spacing
func TestINIWrapper_Configure_InvalidType(t *testing.T) {
	wrapper := NewINIWrapper()

	// Configure with non-boolean value - should be ignored
	err := wrapper.Configure(map[string]interface{}{
		"use_spacing": "not a bool",
	})
	if err != nil {
		t.Errorf("Configure with invalid type should not fail, got: %v", err)
	}

	// Should keep default value (true)
	if !wrapper.parser.useSpacing {
		t.Errorf("Expected useSpacing to remain true after invalid type")
	}
}

// TestINIWrapper_Configure_Default tests that default is true
func TestINIWrapper_Configure_Default(t *testing.T) {
	wrapper := NewINIWrapper()

	if !wrapper.parser.useSpacing {
		t.Errorf("Expected default useSpacing to be true, got false")
	}

	// Check default comment chars
	expectedCommentChars := []byte{'#', ';'}
	if len(wrapper.parser.commentChars) != len(expectedCommentChars) {
		t.Errorf("Expected default commentChars length %d, got %d", len(expectedCommentChars), len(wrapper.parser.commentChars))
	}
	for i, ch := range expectedCommentChars {
		if wrapper.parser.commentChars[i] != ch {
			t.Errorf("Expected commentChars[%d] = %c, got %c", i, ch, wrapper.parser.commentChars[i])
		}
	}

	// Check default delimiter
	if wrapper.parser.delimiter != '=' {
		t.Errorf("Expected default delimiter '=', got '%c'", wrapper.parser.delimiter)
	}
}

// TestINIWrapper_Configure_PreservesExistingDelimiters tests that existing delimiters are preserved
func TestINIWrapper_Configure_PreservesExistingDelimiters(t *testing.T) {
	initialData := `[section]
nospaceskey=value1
spaceskey = value2
mixed=  value3
`

	wrapper := NewINIWrapper()

	// Configure with use_spacing false
	err := wrapper.Configure(map[string]interface{}{
		"use_spacing": false,
	})
	if err != nil {
		t.Fatalf("Configure failed: %v", err)
	}

	// Parse
	data, err := wrapper.Parse([]byte(initialData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Modify existing values
	section := data["section"].(map[string]interface{})
	section["nospaceskey"] = "newvalue1"
	section["spaceskey"] = "newvalue2"
	section["mixed"] = "newvalue3"

	// Serialize
	var buf bytes.Buffer
	err = wrapper.Serialize(data, &buf)
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	result := buf.String()

	// Existing delimiters should be preserved even with use_spacing false
	if !strings.Contains(result, "nospaceskey=newvalue1") {
		t.Errorf("Expected 'nospaceskey=newvalue1' (no spaces) in output, got:\n%s", result)
	}
	if !strings.Contains(result, "spaceskey = newvalue2") {
		t.Errorf("Expected 'spaceskey = newvalue2' (with spaces) in output, got:\n%s", result)
	}
}

// TestINIWrapper_Configure_CommentChars tests custom comment character configuration
func TestINIWrapper_Configure_CommentChars(t *testing.T) {
	tests := []struct {
		name         string
		commentChars string
		inputData    string
		expectParsed bool
		expectKey    string
	}{
		{
			name:         "default hash comment",
			commentChars: "#;",
			inputData:    "# this is a comment\nkey=value\n",
			expectParsed: true,
			expectKey:    "key",
		},
		{
			name:         "default semicolon comment",
			commentChars: "#;",
			inputData:    "; this is a comment\nkey=value\n",
			expectParsed: true,
			expectKey:    "key",
		},
		{
			name:         "custom double-slash comment",
			commentChars: "/",
			inputData:    "// this is a comment\nkey=value\n",
			expectParsed: false, // First slash is comment, second slash becomes part of comment text
			expectKey:    "",
		},
		{
			name:         "only hash as comment",
			commentChars: "#",
			inputData:    "; key=value\nrealkey=realvalue\n",
			expectParsed: true,
			expectKey:    "; key", // Semicolon line is treated as a key since ; is not a comment char
		},
		{
			name:         "multiple custom comment chars",
			commentChars: "#;!",
			inputData:    "! this is a comment\nkey=value\n",
			expectParsed: true,
			expectKey:    "key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(subT *testing.T) {
			wrapper := NewINIWrapper()

			err := wrapper.Configure(map[string]interface{}{
				"comment_chars": tt.commentChars,
			})
			if err != nil {
				subT.Fatalf("Configure failed: %v", err)
			}

			data, err := wrapper.Parse([]byte(tt.inputData))
			if err != nil {
				subT.Fatalf("Parse failed: %v", err)
			}

			rootSection := data[""]
			if tt.expectParsed {
				if rootSection == nil {
					subT.Fatalf("Expected root section to exist")
				}
				section := rootSection.(map[string]interface{})
				if _, exists := section[tt.expectKey]; !exists {
					subT.Errorf("Expected key '%s' to exist in parsed data, keys: %v", tt.expectKey, section)
				}
			}
		})
	}
}

// TestINIWrapper_Configure_Delimiter tests custom delimiter configuration
func TestINIWrapper_Configure_Delimiter(t *testing.T) {
	tests := []struct {
		name      string
		delimiter string
		inputData string
		expectKey string
		expectVal string
	}{
		{
			name:      "default equals delimiter",
			delimiter: "=",
			inputData: "[section]\nkey=value\n",
			expectKey: "key",
			expectVal: "value",
		},
		{
			name:      "colon delimiter",
			delimiter: ":",
			inputData: "[section]\nkey:value\n",
			expectKey: "key",
			expectVal: "value",
		},
		{
			name:      "space delimiter",
			delimiter: " ",
			inputData: "[section]\nkey value\n",
			expectKey: "key",
			expectVal: "value",
		},
		{
			name:      "custom delimiter preserves spacing",
			delimiter: ":",
			inputData: "[section]\nkey : value\n",
			expectKey: "key",
			expectVal: "value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(subT *testing.T) {
			wrapper := NewINIWrapper()

			err := wrapper.Configure(map[string]interface{}{
				"delimiter": tt.delimiter,
			})
			if err != nil {
				subT.Fatalf("Configure failed: %v", err)
			}

			data, err := wrapper.Parse([]byte(tt.inputData))
			if err != nil {
				subT.Fatalf("Parse failed: %v", err)
			}

			section := data["section"].(map[string]interface{})
			if section[tt.expectKey] != tt.expectVal {
				subT.Errorf("Expected %s=%s, got %s=%v", tt.expectKey, tt.expectVal, tt.expectKey, section[tt.expectKey])
			}
		})
	}
}

// TestINIWrapper_Configure_DelimiterInOutput tests that custom delimiter is used in output
func TestINIWrapper_Configure_DelimiterInOutput(t *testing.T) {
	wrapper := NewINIWrapper()

	// Configure with colon delimiter
	err := wrapper.Configure(map[string]interface{}{
		"delimiter": ":",
	})
	if err != nil {
		t.Fatalf("Configure failed: %v", err)
	}

	// Create new content
	data := map[string]interface{}{
		"section": map[string]interface{}{
			"newkey": "newvalue",
		},
	}

	var buf bytes.Buffer
	err = wrapper.Serialize(data, &buf)
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	result := buf.String()

	// New keys should use the custom delimiter
	if !strings.Contains(result, "newkey:newvalue") && !strings.Contains(result, "newkey : newvalue") {
		t.Errorf("Expected 'newkey:newvalue' or 'newkey : newvalue' in output, got:\n%s", result)
	}
}

// TestINIWrapper_Configure_CombinedOptions tests multiple options together
func TestINIWrapper_Configure_CombinedOptions(t *testing.T) {
	wrapper := NewINIWrapper()

	err := wrapper.Configure(map[string]interface{}{
		"use_spacing":   false,
		"comment_chars": "!",
		"delimiter":     ":",
	})
	if err != nil {
		t.Fatalf("Configure failed: %v", err)
	}

	inputData := `! This is a comment
[section]
key:value
`

	data, err := wrapper.Parse([]byte(inputData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Add new key
	section := data["section"].(map[string]interface{})
	section["newkey"] = "newvalue"

	var buf bytes.Buffer
	err = wrapper.Serialize(data, &buf)
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	result := buf.String()

	// Should use colon without spaces (use_spacing: false)
	if !strings.Contains(result, "newkey:newvalue") {
		t.Errorf("Expected 'newkey:newvalue' (no spaces, custom delimiter) in output, got:\n%s", result)
	}

	// Should preserve comment with !
	if !strings.Contains(result, "! This is a comment") {
		t.Errorf("Expected comment with '!' to be preserved, got:\n%s", result)
	}
}

// TestINIWrapper_Configure_InvalidValues tests handling of invalid configuration values
func TestINIWrapper_Configure_InvalidValues(t *testing.T) {
	tests := []struct {
		name    string
		options map[string]interface{}
	}{
		{
			name: "empty comment_chars string",
			options: map[string]interface{}{
				"comment_chars": "",
			},
		},
		{
			name: "empty delimiter string",
			options: map[string]interface{}{
				"delimiter": "",
			},
		},
		{
			name: "wrong type for comment_chars",
			options: map[string]interface{}{
				"comment_chars": 123,
			},
		},
		{
			name: "wrong type for delimiter",
			options: map[string]interface{}{
				"delimiter": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(subT *testing.T) {
			wrapper := NewINIWrapper()

			// Store original values
			origCommentChars := make([]byte, len(wrapper.parser.commentChars))
			copy(origCommentChars, wrapper.parser.commentChars)
			origDelimiter := wrapper.parser.delimiter

			err := wrapper.Configure(tt.options)
			if err != nil {
				subT.Errorf("Configure should not fail on invalid values, got: %v", err)
			}

			// Verify that invalid values don't change the defaults
			if len(origCommentChars) > 0 && tt.options["comment_chars"] != nil {
				// Comment chars should remain unchanged
				if len(wrapper.parser.commentChars) != len(origCommentChars) {
					subT.Errorf("commentChars should remain unchanged with invalid input")
				}
			}

			if tt.options["delimiter"] != nil {
				// Delimiter should remain unchanged
				if wrapper.parser.delimiter != origDelimiter {
					subT.Errorf("delimiter should remain unchanged with invalid input")
				}
			}
		})
	}
}
