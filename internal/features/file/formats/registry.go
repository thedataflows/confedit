package formats

import "fmt"

// Registry manages all available format parsers
type Registry struct {
	parsers map[string]Parser
}

// NewRegistry creates a new format parser registry
func NewRegistry() *Registry {
	return &Registry{
		parsers: make(map[string]Parser),
	}
}

// Register adds a parser for a specific format
func (r *Registry) Register(format string, parser Parser) {
	r.parsers[format] = parser
}

// Get retrieves a parser for the specified format
func (r *Registry) Get(format string) (Parser, error) {
	parser, exists := r.parsers[format]
	if !exists {
		return nil, fmt.Errorf("no parser registered for format: %s", format)
	}
	return parser, nil
}

// Has checks if a parser is registered for the format
func (r *Registry) Has(format string) bool {
	_, exists := r.parsers[format]
	return exists
}

// Formats returns all registered format names
func (r *Registry) Formats() []string {
	formats := make([]string, 0, len(r.parsers))
	for format := range r.parsers {
		formats = append(formats, format)
	}
	return formats
}
