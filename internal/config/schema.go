package config

import (
	"embed"
	"fmt"
	"os"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"github.com/thedataflows/confedit/internal/types"
	log "github.com/thedataflows/go-lib-log"
)

//go:embed schema.cue
var schemaFS embed.FS

// SchemaValidator validates configuration against CUE schema
type SchemaValidator struct {
	ctx    *cue.Context
	schema cue.Value
}

// NewSchemaValidator creates a new schema validator with optional custom schema file
func NewSchemaValidator(schemaFilePath ...string) (*SchemaValidator, error) {
	ctx := cuecontext.New()

	var schemaData []byte
	var err error

	// Try to use custom schema file if provided
	if len(schemaFilePath) > 0 && schemaFilePath[0] != "" {
		if _, err := os.Stat(schemaFilePath[0]); err == nil {
			schemaData, err = os.ReadFile(schemaFilePath[0])
			if err != nil {
				log.Logger().Warn().Err(err).Str("schema", schemaFilePath[0]).Msg("Failed to read custom schema, falling back to embedded schema")
			} else {
				log.Debugf("schema", "Using custom schema '%s'", schemaFilePath[0])
			}
		} else {
			log.Logger().Warn().Err(err).Str("schema", schemaFilePath[0]).Msg("Custom schema not found or not readable, falling back to embedded schema")
		}
	}

	// Fall back to embedded schema if custom schema failed or wasn't provided
	if len(schemaData) == 0 {
		schemaData, err = schemaFS.ReadFile("schema.cue")
		if err != nil {
			return nil, fmt.Errorf("read embedded schema: %w", err)
		}
		log.Debug("schema", "Using embedded schema")
	}

	// Compile the schema
	schema := ctx.CompileBytes(schemaData)
	if err := schema.Err(); err != nil {
		return nil, fmt.Errorf("compile schema: %w", err)
	}

	return &SchemaValidator{
		ctx:    ctx,
		schema: schema,
	}, nil
}

// ValidateConfig validates a SystemConfig against the schema
func (sv *SchemaValidator) ValidateConfig(config *types.SystemConfig) error {
	// Convert config to CUE value
	configValue := sv.ctx.Encode(config)
	if err := configValue.Err(); err != nil {
		return fmt.Errorf("encode config: %w", err)
	}

	// Get the SystemConfig schema
	systemConfigSchema := sv.schema.LookupPath(cue.ParsePath("#SystemConfig"))
	if err := systemConfigSchema.Err(); err != nil {
		return fmt.Errorf("find SystemConfig schema: %w", err)
	}

	// Unify the config with the schema
	unified := configValue.Unify(systemConfigSchema)
	if err := unified.Err(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Validate that the unified result is complete and concrete
	if err := unified.Validate(cue.Concrete(true), cue.Final()); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}

// ValidateRawConfig validates raw configuration data against the schema
func (sv *SchemaValidator) ValidateRawConfig(data []byte) error {
	// Parse the raw data as CUE
	configValue := sv.ctx.CompileBytes(data)
	if err := configValue.Err(); err != nil {
		return fmt.Errorf("parse config data: %w", err)
	}

	// Get the SystemConfig schema
	systemConfigSchema := sv.schema.LookupPath(cue.ParsePath("#SystemConfig"))
	if err := systemConfigSchema.Err(); err != nil {
		return fmt.Errorf("failed to find SystemConfig schema: %w", err)
	}

	// Unify the config with the schema
	unified := configValue.Unify(systemConfigSchema)
	if err := unified.Err(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Validate that the unified result is complete and concrete
	if err := unified.Validate(cue.Concrete(true), cue.Final()); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}
