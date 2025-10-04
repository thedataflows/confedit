package toml

import (
	"io"

	"github.com/pelletier/go-toml/v2"
	"github.com/thedataflows/confedit/internal/features/file/formats"
)

// Parser implements the formats.Parser interface for TOML files
type Parser struct{}

// New creates a new TOML parser
func New() formats.Parser {
	return &Parser{}
}

func (p *Parser) Unmarshal(data []byte) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := toml.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (p *Parser) Marshal(data map[string]interface{}, writer io.Writer) error {
	// Use the stable API for reliable serialization
	// Comments will be lost, but the data structure will be preserved correctly
	encoded, err := toml.Marshal(data)
	if err != nil {
		return err
	}

	_, err = writer.Write(encoded)
	return err
}

// Verify that Parser implements the interface at compile time
var _ formats.Parser = (*Parser)(nil)
