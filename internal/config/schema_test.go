package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/thedataflows/confedit/internal/features/file"
	"github.com/thedataflows/confedit/internal/types"
)

type SchemaTestSuite struct {
	suite.Suite
	validator *SchemaValidator
}

func TestSchemaTestSuite(t *testing.T) {
	suite.Run(t, new(SchemaTestSuite))
}

func (s *SchemaTestSuite) SetupTest() {
	validator, err := NewSchemaValidator()
	require.NoError(s.T(), err)
	s.validator = validator
}

func (s *SchemaTestSuite) TestNewSchemaValidator() {
	validator, err := NewSchemaValidator()
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), validator)
}

func (s *SchemaTestSuite) TestValidateConfig_ValidFileConfig() {
	validConfig := &types.SystemConfig{
		Targets: []types.AnyTarget{
			&file.Target{
				Name: "test-config",
				Type: types.TYPE_FILE,
				Config: &file.Config{
					Path:   "/etc/test.conf",
					Format: "ini",
					Content: map[string]interface{}{
						"section1": map[string]interface{}{
							"key1": "value1",
						},
					},
				},
			},
		},
	}

	err := s.validator.ValidateConfig(validConfig)
	assert.NoError(s.T(), err)
}

func (s *SchemaTestSuite) TestValidateConfig_InvalidFileConfig_MissingPath() {
	invalidConfig := &types.SystemConfig{
		Targets: []types.AnyTarget{
			&file.Target{
				Name: "test-config",
				Type: types.TYPE_FILE,
				Config: &file.Config{
					// Missing required Path field
					Format: "ini",
					Content: map[string]interface{}{
						"section1": map[string]interface{}{
							"key1": "value1",
						},
					},
				},
			},
		},
	}

	err := s.validator.ValidateConfig(invalidConfig)
	assert.Error(s.T(), err)
}

func (s *SchemaTestSuite) TestValidateConfig_InvalidFileConfig_MissingFormat() {
	invalidConfig := &types.SystemConfig{
		Targets: []types.AnyTarget{
			&file.Target{
				Name: "test-config",
				Type: types.TYPE_FILE,
				Config: &file.Config{
					Path: "/etc/test.conf",
					// Missing required Format field
					Content: map[string]interface{}{
						"section1": map[string]interface{}{
							"key1": "value1",
						},
					},
				},
			},
		},
	}

	err := s.validator.ValidateConfig(invalidConfig)
	assert.Error(s.T(), err)
}

func (s *SchemaTestSuite) TestValidateConfig_InvalidType() {
	invalidConfig := &types.SystemConfig{
		Targets: []types.AnyTarget{
			&file.Target{
				Name: "test-config",
				Type: "invalid-type", // Invalid type
				Config: &file.Config{
					Path:    "/etc/test.conf",
					Format:  "ini",
					Content: map[string]interface{}{},
				},
			},
		},
	}

	err := s.validator.ValidateConfig(invalidConfig)
	assert.Error(s.T(), err)
}

func (s *SchemaTestSuite) TestValidateConfig_ValidWithVariables() {
	validConfig := &types.SystemConfig{
		Targets: []types.AnyTarget{
			&file.Target{
				Name: "test-config",
				Type: types.TYPE_FILE,
				Config: &file.Config{
					Path:   "/etc/test.conf",
					Format: "ini",
					Content: map[string]interface{}{
						"section1": map[string]interface{}{
							"key1": "value1",
						},
					},
				},
			},
		},
		Variables: map[string]interface{}{
			"VAR1": "value1",
			"VAR2": 42,
			"VAR3": true,
		},
	}

	err := s.validator.ValidateConfig(validConfig)
	assert.NoError(s.T(), err)
}

func (s *SchemaTestSuite) TestValidateConfig_ValidWithHooks() {
	validConfig := &types.SystemConfig{
		Targets: []types.AnyTarget{
			&file.Target{
				Name: "test-config",
				Type: types.TYPE_FILE,
				Config: &file.Config{
					Path:   "/etc/test.conf",
					Format: "ini",
					Content: map[string]interface{}{
						"section1": map[string]interface{}{
							"key1": "value1",
						},
					},
				},
			},
		},
		Hooks: &types.Hooks{
			PreApply: []string{
				"echo 'before'",
				"systemctl stop service",
			},
			PostApply: []string{
				"echo 'after'",
				"systemctl start service",
			},
		},
	}

	err := s.validator.ValidateConfig(validConfig)
	assert.NoError(s.T(), err)
}

func (s *SchemaTestSuite) TestValidateConfig_EmptyConfig() {
	emptyConfig := &types.SystemConfig{
		Targets: []types.AnyTarget{},
	}

	err := s.validator.ValidateConfig(emptyConfig)
	assert.NoError(s.T(), err)
}

func (s *SchemaTestSuite) TestValidateConfig_NilConfig() {
	err := s.validator.ValidateConfig(nil)
	assert.Error(s.T(), err)
}

func (s *SchemaTestSuite) TestValidateConfig_ComplexFileConfig() {
	validConfig := &types.SystemConfig{
		Targets: []types.AnyTarget{
			&file.Target{
				Name: "complex-config",
				Type: types.TYPE_FILE,
				Metadata: map[string]interface{}{
					"description": "Complex configuration",
					"version":     "1.0",
				},
				Config: &file.Config{
					Path:   "/etc/complex.conf",
					Format: "ini",
					Owner:  "root",
					Group:  "root",
					Mode:   "0644",
					Backup: true,
					Content: map[string]interface{}{
						"section1": map[string]interface{}{
							"key1": "value1",
							"key2": "42",   // INI values should be strings
							"key3": "true", // INI values should be strings
						},
						"section2": map[string]interface{}{
							"nested_key": "value",
						},
					},
					Options: map[string]interface{}{
						"use_spacing": false,
					},
				},
			},
		},
	}

	err := s.validator.ValidateConfig(validConfig)
	assert.NoError(s.T(), err)
}

func (s *SchemaTestSuite) TestValidateConfig_MultipleTargets() {
	validConfig := &types.SystemConfig{
		Targets: []types.AnyTarget{
			&file.Target{
				Name: "target1",
				Type: types.TYPE_FILE,
				Config: &file.Config{
					Path:   "/etc/target1.conf",
					Format: "ini",
					Content: map[string]interface{}{
						"section1": map[string]interface{}{
							"key1": "value1",
						},
					},
				},
			},
			&file.Target{
				Name: "target2",
				Type: types.TYPE_FILE,
				Config: &file.Config{
					Path:   "/etc/target2.conf",
					Format: "yaml", // YAML can have more flexible content
					Content: map[string]interface{}{
						"config": map[string]interface{}{
							"setting": "value",
						},
					},
				},
			},
		},
	}

	err := s.validator.ValidateConfig(validConfig)
	assert.NoError(s.T(), err)
}

func (s *SchemaTestSuite) TestValidateConfig_InvalidFormat() {
	invalidConfig := &types.SystemConfig{
		Targets: []types.AnyTarget{
			&file.Target{
				Name: "test-config",
				Type: types.TYPE_FILE,
				Config: &file.Config{
					Path:   "/etc/test.conf",
					Format: "invalid-format", // Invalid format
					Content: map[string]interface{}{
						"section1": map[string]interface{}{
							"key1": "value1",
						},
					},
				},
			},
		},
	}

	err := s.validator.ValidateConfig(invalidConfig)
	assert.Error(s.T(), err)
}

func TestSchemaValidator_RawConfig(t *testing.T) {
	validator, err := NewSchemaValidator()
	if err != nil {
		t.Fatalf("Failed to create schema validator: %v", err)
	}

	// Test valid raw configuration
	validRawConfig := `{
		"targets": [
			{
				"name": "test-config",
				"type": "file",
				"config": {
					"path": "/etc/test.conf",
					"format": "ini",
					"content": {
						"section1": {
							"key1": "value1"
						}
					}
				}
			}
		]
	}`

	err = validator.ValidateRawConfig([]byte(validRawConfig))
	if err != nil {
		t.Errorf("Valid raw config should pass validation: %v", err)
	}

	// Test invalid raw configuration
	invalidRawConfig := `{
		"targets": [
			{
				"name": "test-config",
				"type": "invalid-type"
			}
		]
	}`

	err = validator.ValidateRawConfig([]byte(invalidRawConfig))
	if err == nil {
		t.Error("Invalid raw config should fail validation")
	}
}

func TestCueConfigLoader_WithCustomSchema(t *testing.T) {
	// Create a temporary schema file for testing
	schemaFile := "/tmp/test_loader_schema.cue"
	schemaContent := `package config
#ConfigTarget: { name: string, type: "file" }
#SystemConfig: { targets: [...#ConfigTarget] }`

	err := os.WriteFile(schemaFile, []byte(schemaContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create temporary schema file: %v", err)
	}
	defer os.Remove(schemaFile)

	// Test loader with custom schema
	loader := NewCueConfigLoader("./testdata", schemaFile)

	// Verify the loader was created (even if schema validation may fail)
	if loader == nil {
		t.Error("Loader should be created even with custom schema")
	}

	// Verify that validator was initialized (or warning was logged)
	// This test mainly checks that the API works correctly
}
