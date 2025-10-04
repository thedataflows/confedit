package yaml

import (
	"io"

	"github.com/goccy/go-yaml"
	"github.com/thedataflows/confedit/internal/features/file/formats"
)

// Parser implements the formats.Parser interface for YAML files
type Parser struct{}

// New creates a new YAML parser
func New() formats.Parser {
	return &Parser{}
}

func (p *Parser) Unmarshal(data []byte) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := yaml.Unmarshal(data, &result)
	return result, err
}

func (p *Parser) Marshal(data map[string]interface{}, writer io.Writer) error {
	bytes, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	_, err = writer.Write(bytes)
	return err
}

// Verify that Parser implements the interface at compile time
var _ formats.Parser = (*Parser)(nil)
