package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeNameFromSource(t1 *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only alphanumeric",
			input:    "abc123",
			expected: "abc123",
		},
		{
			name:     "path with slashes",
			input:    "/etc/nginx/nginx.conf",
			expected: "etc-nginx-nginx-conf",
		},
		{
			name:     "path with dots",
			input:    "config.yaml",
			expected: "config-yaml",
		},
		{
			name:     "multiple consecutive special chars",
			input:    "test...config___file",
			expected: "test-config-file",
		},
		{
			name:     "leading and trailing special chars",
			input:    "..config..",
			expected: "config",
		},
	}

	for _, tt := range tests {
		t1.Run(tt.name, func(t2 *testing.T) {
			result := normalizeNameFromSource(tt.input)
			assert.Equal(t2, tt.expected, result)
		})
	}
}

func TestDetectFormat(t1 *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"ini file", "config.ini", "ini"},
		{"conf file", "app.conf", "ini"},
		{"yaml file", "config.yaml", "yaml"},
		{"yml file", "config.yml", "yaml"},
		{"toml file", "config.toml", "toml"},
		{"json file", "config.json", "json"},
		{"xml file", "config.xml", "xml"},
		{"unknown extension", "config.txt", ""},
		{"no extension", "config", ""},
	}

	cmd := &GenerateCmd{}
	for _, tt := range tests {
		t1.Run(tt.name, func(t2 *testing.T) {
			result := cmd.detectFormat(tt.path)
			assert.Equal(t2, tt.expected, result)
		})
	}
}

func TestGenerateCUEOutput(t *testing.T) {
	cmd := &GenerateCmd{
		Name: "test-config",
		Type: "file",
	}

	tests := []struct {
		name     string
		diff     map[string]interface{}
		expected string
	}{
		{
			name: "simple key-value pairs",
			diff: map[string]interface{}{
				"host": "localhost",
				"port": 8080,
				"ssl":  true,
			},
			expected: `package config

{
	targets: [{
		config: {
			content: {
				host: "localhost"
				port: 8080
				ssl:  true
			}
			format: "toml"
			path:   "test.toml"
		}
		name: "test-config"
		type: "file"
	}]
}`,
		},
		{
			name: "nested sections",
			diff: map[string]interface{}{
				"server": map[string]interface{}{
					"host": "localhost",
					"port": 8080,
				},
				"database": map[string]interface{}{
					"name": "myapp",
					"user": "admin",
				},
			},
			expected: `package config

{
	targets: [{
		config: {
			content: {
				database: {
					name: "myapp"
					user: "admin"
				}
				server: {
					host: "localhost"
					port: 8080
				}
			}
			format: "toml"
			path:   "test.toml"
		}
		name: "test-config"
		type: "file"
	}]
}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(sub *testing.T) {
			result := cmd.generateSimpleCUE("test.toml", tc.diff)
			assert.Equal(sub, tc.expected, result)
		})
	}
}

func TestValuesEqual(t *testing.T) {
	cmd := &GenerateCmd{}

	tests := []struct {
		name     string
		a        interface{}
		b        interface{}
		expected bool
	}{
		{"equal strings", "hello", "hello", true},
		{"different strings", "hello", "world", false},
		{"equal ints", 42, 42, true},
		{"different ints", 42, 43, false},
		{"equal bools", true, true, true},
		{"different bools", true, false, false},
		{"nil values", nil, nil, true},
		{"nil vs string", nil, "hello", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(sub *testing.T) {
			result := cmd.valuesEqual(tc.a, tc.b)
			assert.Equal(sub, tc.expected, result)
		})
	}
}
