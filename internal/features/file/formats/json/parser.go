package jsonformat

import (
	"io"

	"github.com/goccy/go-json"
	"github.com/thedataflows/confedit/internal/features/file/formats"
)

// Parser implements the formats.Parser interface for JSON files
type Parser struct{}

// New creates a new JSON parser
func New() formats.Parser {
	return &Parser{}
}

func (p *Parser) Unmarshal(data []byte) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := json.Unmarshal(data, &result)
	return result, err
}

func (p *Parser) Marshal(data map[string]interface{}, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// Verify that Parser implements the interface at compile time
var _ formats.Parser = (*Parser)(nil)
