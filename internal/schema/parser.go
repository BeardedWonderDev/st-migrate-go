package schema

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Parser defines how to decode a schema version into an in-memory Spec.
type Parser interface {
	Parse(data []byte) (*Spec, error)
}

// Registry maps schema versions to parsers.
type Registry struct {
	parsers map[int]Parser
}

// NewRegistry builds an empty registry.
func NewRegistry() *Registry {
	return &Registry{parsers: map[int]Parser{}}
}

// Register attaches a parser for a schema version.
func (r *Registry) Register(version int, parser Parser) {
	r.parsers[version] = parser
}

// Parse uses the registered parser for the schema version in the YAML payload.
// If the version is missing or zero, schema version 1 is assumed.
func (r *Registry) Parse(data []byte) (*Spec, error) {
	var meta struct {
		Version int `yaml:"version"`
	}
	if err := yaml.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("parse schema metadata: %w", err)
	}
	schemaVersion := meta.Version
	if schemaVersion == 0 {
		schemaVersion = 1
	}
	parser, ok := r.parsers[schemaVersion]
	if !ok {
		return nil, fmt.Errorf("unsupported schema version %d", schemaVersion)
	}
	return parser.Parse(data)
}

// DefaultRegistry returns a registry populated with built-in schema parsers.
func DefaultRegistry() *Registry {
	r := NewRegistry()
	r.Register(1, V1Parser{})
	return r
}
