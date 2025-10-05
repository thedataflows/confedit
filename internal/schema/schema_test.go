package schema

import (
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

func (s *SchemaTestSuite) TestValidate_ValidFileConfig() {
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

	err := s.validator.Validate(validConfig)
	assert.NoError(s.T(), err)
}

func (s *SchemaTestSuite) TestValidate_InvalidFileConfig_MissingPath() {
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

	err := s.validator.Validate(invalidConfig)
	assert.Error(s.T(), err)
}

func (s *SchemaTestSuite) TestValidate_InvalidFileConfig_MissingFormat() {
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

	err := s.validator.Validate(invalidConfig)
	assert.Error(s.T(), err)
}

func (s *SchemaTestSuite) TestValidate_InvalidType() {
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

	err := s.validator.Validate(invalidConfig)
	assert.Error(s.T(), err)
}

func (s *SchemaTestSuite) TestValidate_ValidWithVariables() {
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

	err := s.validator.Validate(validConfig)
	assert.NoError(s.T(), err)
}

func (s *SchemaTestSuite) TestValidate_ValidWithHooks() {
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

	err := s.validator.Validate(validConfig)
	assert.NoError(s.T(), err)
}

func (s *SchemaTestSuite) TestValidate_EmptyConfig() {
	emptyConfig := &types.SystemConfig{
		Targets: []types.AnyTarget{},
	}

	err := s.validator.Validate(emptyConfig)
	assert.NoError(s.T(), err)
}

func (s *SchemaTestSuite) TestValidate_NilConfig() {
	err := s.validator.Validate(nil)
	assert.Error(s.T(), err)
}

func (s *SchemaTestSuite) TestValidate_ComplexFileConfig() {
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

	err := s.validator.Validate(validConfig)
	assert.NoError(s.T(), err)
}

func (s *SchemaTestSuite) TestValidate_MultipleTargets() {
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

	err := s.validator.Validate(validConfig)
	assert.NoError(s.T(), err)
}

func (s *SchemaTestSuite) TestValidate_InvalidFormat() {
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

	err := s.validator.Validate(invalidConfig)
	assert.Error(s.T(), err)
}

func TestSchemaValidator_RawConfig(t *testing.T) {
	validator, err := NewSchemaValidator()
	if err != nil {
		t.Fatalf("create schema validator: %v", err)
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

	err = validator.ValidateRaw([]byte(validRawConfig))
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

	err = validator.ValidateRaw([]byte(invalidRawConfig))
	if err == nil {
		t.Error("Invalid raw config should fail validation")
	}
}
