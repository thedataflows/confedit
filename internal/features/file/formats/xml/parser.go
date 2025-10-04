package xml

import (
	"encoding/xml"
	"io"

	"github.com/thedataflows/confedit/internal/features/file/formats"
)

// Parser implements the formats.Parser interface for XML files
type Parser struct{}

// New creates a new XML parser
func New() formats.Parser {
	return &Parser{}
}

func (p *Parser) Unmarshal(data []byte) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := xml.Unmarshal(data, &result)
	return result, err
}

func (p *Parser) Marshal(data map[string]interface{}, writer io.Writer) error {
	encoder := xml.NewEncoder(writer)
	encoder.Indent("", "  ")
	err := encoder.Encode(data)
	if err != nil {
		return err
	}
	return encoder.Flush()
}

// Verify that Parser implements the interface at compile time
var _ formats.Parser = (*Parser)(nil)
