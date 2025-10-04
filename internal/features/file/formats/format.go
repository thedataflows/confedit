package formats

import "io"

// Parser defines the interface for configuration file format parsers
// Each format (ini, yaml, toml, json, xml) implements this interface
type Parser interface {
	// Unmarshal parses the configuration data from bytes into a map
	Unmarshal(data []byte) (map[string]interface{}, error)

	// Marshal serializes the configuration map into the format and writes to the writer
	Marshal(data map[string]interface{}, writer io.Writer) error
}

// ConfigurableParser extends Parser with configuration capabilities
// Some parsers (like INI) need format-specific options
type ConfigurableParser interface {
	Parser

	// Configure sets parser-specific options (e.g., INI delimiter, comment chars)
	Configure(options map[string]interface{}) error
}
