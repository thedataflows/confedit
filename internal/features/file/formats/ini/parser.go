package ini

import (
	"io"

	"github.com/thedataflows/confedit/internal/features/file/formats"
	"github.com/thedataflows/confedit/internal/features/file/formats/ini/iniparser"
)

// Parser implements the formats.Parser interface for INI files
// It wraps the INI parser implementation for proper format handling
type Parser struct {
	wrapper *iniparser.INIWrapper
}

// New creates a new INI parser
func New() formats.Parser {
	return &Parser{
		wrapper: iniparser.NewINIWrapper(),
	}
}

// Unmarshal parses INI data and returns a nested map structure
// Structure: map[section]map[key]value
// Root keys (no section) are stored under "" (empty string) section
func (p *Parser) Unmarshal(data []byte) (map[string]interface{}, error) {
	return p.wrapper.Unmarshal(data)
}

// Marshal writes the map structure back to INI format
// Preserves formatting and structure from previous parse if available
func (p *Parser) Marshal(data map[string]interface{}, writer io.Writer) error {
	return p.wrapper.Marshal(data, writer)
}

// Configure implements ConfigurableParser to accept INI-specific options
// Supported options:
//   - use_spacing (bool): Controls delimiter formatting for new keys
//     When true (default), new keys are written as "key = value"
//     When false, new keys are written as "key=value"
//   - comment_chars (string): Characters to recognize as comment prefixes
//     Default is "#;" (both # and ; are recognized)
//   - delimiter (string): Key-value delimiter character (default is "=")
func (p *Parser) Configure(options map[string]interface{}) error {
	return p.wrapper.Configure(options)
}

// Verify that Parser implements both interfaces at compile time
var (
	_ formats.Parser             = (*Parser)(nil)
	_ formats.ConfigurableParser = (*Parser)(nil)
)
