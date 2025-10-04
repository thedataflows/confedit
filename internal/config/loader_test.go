package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/thedataflows/confedit/internal/features/file"
)

type ConfigLoaderTestSuite struct {
	suite.Suite
	tempDir string
}

func TestConfigLoaderTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigLoaderTestSuite))
}

func (s *ConfigLoaderTestSuite) SetupTest() {
	tempDir, err := os.MkdirTemp("", "config_loader_test")
	require.NoError(s.T(), err)
	s.tempDir = tempDir
}

func (s *ConfigLoaderTestSuite) TearDownTest() {
	if s.tempDir != "" {
		os.RemoveAll(s.tempDir)
	}
}

func (s *ConfigLoaderTestSuite) TestNewCueConfigLoader() {
	tests := []struct {
		name           string
		configPath     string
		schemaFilePath []string
		expectError    bool
	}{
		{
			name:       "valid config path without schema",
			configPath: "/some/path",
		},
		{
			name:           "valid config path with schema",
			configPath:     "/some/path",
			schemaFilePath: []string{"schema.cue"},
		},
		{
			name:           "empty schema path",
			configPath:     "/some/path",
			schemaFilePath: []string{""},
		},
		{
			name:           "multiple schema files",
			configPath:     "/some/path",
			schemaFilePath: []string{"schema1.cue", "schema2.cue"},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			loader := NewCueConfigLoader(tt.configPath, tt.schemaFilePath...)

			assert.NotNil(s.T(), loader, "expected loader to be created")
			assert.Equal(s.T(), tt.configPath, loader.configPath, "configPath should match")
			assert.NotNil(s.T(), loader.ctx, "expected CUE context to be initialized")
			assert.NotNil(s.T(), loader.validator, "expected validator to be initialized")
		})
	}
}

func (s *ConfigLoaderTestSuite) TestCollectCueFiles() {
	// Create test files
	file1 := filepath.Join(s.tempDir, "01-config.cue")
	file2 := filepath.Join(s.tempDir, "02-config.cue")
	subDir := filepath.Join(s.tempDir, "subdir")
	file3 := filepath.Join(subDir, "03-config.cue")
	nonCueFile := filepath.Join(s.tempDir, "other.txt")

	err := os.WriteFile(file1, []byte("package config\ntargets: []"), 0644)
	require.NoError(s.T(), err)
	err = os.WriteFile(file2, []byte("package config\ntargets: []"), 0644)
	require.NoError(s.T(), err)
	err = os.MkdirAll(subDir, 0755)
	require.NoError(s.T(), err)
	err = os.WriteFile(file3, []byte("package config\ntargets: []"), 0644)
	require.NoError(s.T(), err)
	err = os.WriteFile(nonCueFile, []byte("not a cue file"), 0644)
	require.NoError(s.T(), err)

	tests := []struct {
		name          string
		configPath    string
		expectedFiles []string
		expectError   bool
	}{
		{
			name:          "single file",
			configPath:    file1,
			expectedFiles: []string{file1},
		},
		{
			name:          "directory with multiple files",
			configPath:    s.tempDir,
			expectedFiles: []string{file1, file2, file3}, // Should be sorted
		},
		{
			name:        "non-existent path",
			configPath:  "/non/existent/path",
			expectError: true,
		},
		{
			name:        "non-cue file",
			configPath:  nonCueFile,
			expectError: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			loader := NewCueConfigLoader(tt.configPath)
			files, err := loader.collectCueFiles()

			if tt.expectError {
				assert.Error(s.T(), err, "expected error")
				return
			}

			require.NoError(s.T(), err, "unexpected error")
			assert.Len(s.T(), files, len(tt.expectedFiles), "file count should match")

			// Check files are sorted and match expected
			for i, expectedFile := range tt.expectedFiles {
				require.True(s.T(), i < len(files), "missing expected file: %s", expectedFile)
				if i < len(files) {
					assert.Equal(s.T(), expectedFile, files[i], "file at index %d should match", i)
				}
			}
		})
	}
}

func TestCueConfigLoader_LoadConfiguration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a simple test CUE file
	configFile := filepath.Join(tmpDir, "test.cue")
	configContent := `package config

targets: [
	{
		name: "test-target"
		type: "file"
		config: {
			path: "/tmp/test.conf"
			format: "ini"
			content: {
				section1: {
					key1: "value1"
				}
			}
		}
	}
]

variables: {
	VAR1: "value1"
	VAR2: "value2"
}
`

	assert.NoError(t, os.WriteFile(configFile, []byte(configContent), 0644))

	tests := []struct {
		name        string
		configPath  string
		expectError bool
	}{
		{
			name:       "valid single file",
			configPath: configFile,
		},
		{
			name:       "valid directory",
			configPath: tmpDir,
		},
		{
			name:        "non-existent path",
			configPath:  "/non/existent/path",
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(subT *testing.T) {
			loader := NewCueConfigLoader(test.configPath)
			config, err := loader.LoadConfiguration()

			if test.expectError {
				assert.Error(subT, err, "expected error")
				return
			}

			assert.NoError(subT, err, "unexpected error")
			assert.True(subT, config != nil, "expected config to be loaded")

			// Check loaded configuration
			assert.Equal(subT, 1, len(config.Targets), "expected 1 target")

			target := config.Targets[0]
			assert.Equal(subT, "test-target", target.GetName(), "target name should match")
			assert.Equal(subT, 2, len(config.Variables), "expected 2 variables")
			assert.Equal(subT, "value1", config.Variables["VAR1"], "VAR1 should match")
		})
	}
}

func TestCueConfigLoader_MergeTargets(t *testing.T) {
	tmpDir := t.TempDir()

	// Create two CUE files with the same target name
	file1 := filepath.Join(tmpDir, "01-config.cue")
	file1Content := `package config

targets: [
	{
		name: "shared-target"
		type: "file"
		config: {
			path: "/tmp/test.conf"
			format: "ini"
			content: {
				section1: {
					key1: "value1"
				}
			}
		}
	}
]
`

	file2 := filepath.Join(tmpDir, "02-config.cue")
	file2Content := `package config

targets: [
	{
		name: "shared-target"
		type: "file"
		config: {
			path: "/tmp/test.conf"
			format: "ini"
			content: {
				section1: {
					key2: "value2"
				}
				section2: {
					key3: "value3"
				}
			}
		}
	}
]
`

	assert.NoError(t, os.WriteFile(file1, []byte(file1Content), 0644))
	assert.NoError(t, os.WriteFile(file2, []byte(file2Content), 0644))

	loader := NewCueConfigLoader(tmpDir)
	config, err := loader.LoadConfiguration()
	assert.NoError(t, err, "unexpected error")

	// Should have only one target (merged)
	assert.Equal(t, 1, len(config.Targets), "expected 1 merged target")

	target := config.Targets[0]
	assert.Equal(t, "shared-target", target.GetName(), "target name should match")

	// Check content is merged in the target's configuration
	fileTarget, ok := target.(*file.Target)
	assert.True(t, ok, "expected target to be a FileConfigTarget")

	content := fileTarget.Config.Content
	assert.True(t, content != nil, "expected content to be present")

	section1, ok := content["section1"].(map[string]interface{})
	assert.True(t, ok, "expected section1 to be a map")

	assert.Equal(t, "value1", section1["key1"], "key1 should match")
	assert.Equal(t, "value2", section1["key2"], "key2 should match")

	section2, ok := content["section2"].(map[string]interface{})
	assert.True(t, ok, "expected section2 to be a map")
	assert.Equal(t, "value3", section2["key3"], "key3 should match")

	// Check source files metadata (this would be set by the loading process)
	// For now, we're just checking that the metadata exists
	metadata := target.GetMetadata()
	assert.True(t, metadata != nil, "expected metadata to be present")
}

func TestCueConfigLoader_ValidateConfiguration(t *testing.T) {
	tmpDir := t.TempDir()

	validConfigFile := filepath.Join(tmpDir, "valid.cue")
	validContent := `package config

targets: [
	{
		name: "test-target"
		type: "file"
		config: {
			path: "/tmp/test.conf"
			format: "ini"
			content: {
				section1: {
					key1: "value1"
				}
			}
		}
	}
]
`

	invalidConfigFile := filepath.Join(tmpDir, "invalid.cue")
	invalidContent := `package config

this is invalid cue syntax
`

	assert.NoError(t, os.WriteFile(validConfigFile, []byte(validContent), 0644))
	assert.NoError(t, os.WriteFile(invalidConfigFile, []byte(invalidContent), 0644))

	tests := []struct {
		name        string
		configPath  string
		expectError bool
	}{
		{
			name:       "valid config",
			configPath: validConfigFile,
		},
		{
			name:        "invalid config",
			configPath:  invalidConfigFile,
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(subT *testing.T) {
			loader := NewCueConfigLoader(test.configPath)
			err := loader.ValidateConfiguration()

			if test.expectError {
				assert.Error(subT, err, "expected validation error for invalid config")
			} else {
				assert.NoError(subT, err, "expected validation to pass for valid config")
			}
		})
	}
}

func TestCueConfigLoader_ValidationEnforced(t *testing.T) {
	// Create a temporary directory with test files
	tempDir := t.TempDir()

	// Create a valid CUE file
	validCueFile := filepath.Join(tempDir, "valid.cue")
	validContent := `package config

targets: [
	{
		name: "test-target"
		type: "file"
		config: {
			path: "/test/path"
			format: "ini"
			content: {
				section: {
					key: "value"
				}
			}
		}
	}
]
`
	assert.NoError(t, os.WriteFile(validCueFile, []byte(validContent), 0644))

	// Create an invalid CUE file (missing required fields)
	invalidCueFile := filepath.Join(tempDir, "invalid.cue")
	invalidContent := `package config

targets: [
	{
		name: "test-target"
		type: "file"
		config: {
			// missing required path field
			format: "ini"
			content: {
				section: {
					key: "value"
				}
			}
		}
	}
]
`
	assert.NoError(t, os.WriteFile(invalidCueFile, []byte(invalidContent), 0644))

	t.Run("valid config passes validation", func(subT *testing.T) {
		loader := NewCueConfigLoader(validCueFile)

		// Load configuration should succeed with validation
		config, err := loader.LoadConfiguration()
		assert.NoError(subT, err, "expected valid config to load successfully")
		assert.True(subT, config != nil, "expected config to be returned")
		assert.Equal(subT, 1, len(config.Targets), "expected 1 target")
	})

	t.Run("invalid config fails validation", func(subT *testing.T) {
		loader := NewCueConfigLoader(invalidCueFile)

		// Load configuration should fail due to validation
		_, err := loader.LoadConfiguration()
		assert.Error(subT, err, "expected invalid config to fail validation")

		// Error should mention validation failure
		errorMsg := err.Error()
		assert.Contains(subT, errorMsg, "validation", "expected error to mention validation failure")
	})

	t.Run("validator always initialized", func(subT *testing.T) {
		// Test without schema path
		loader1 := NewCueConfigLoader(validCueFile)
		assert.True(subT, loader1.validator != nil, "expected validator to be initialized even without schema path")

		// Test with empty schema path
		loader2 := NewCueConfigLoader(validCueFile, "")
		assert.True(subT, loader2.validator != nil, "expected validator to be initialized with empty schema path")

		// Test with non-existent schema path (should fall back to embedded)
		loader3 := NewCueConfigLoader(validCueFile, "/non/existent/path")
		assert.True(subT, loader3.validator != nil, "expected validator to be initialized with fallback to embedded schema")
	})
}
