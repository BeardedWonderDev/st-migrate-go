package schema

// Action represents a single role/permission operation from a migration spec.
type Action struct {
	Role   string   `yaml:"role"`
	Ensure string   `yaml:"ensure"`
	Add    []string `yaml:"add"`
	Remove []string `yaml:"remove"`
}

// Spec is the parsed YAML document for one migration direction.
// Version is the schema version (not the migration filename version).
type Spec struct {
	Version int      `yaml:"version"`
	Actions []Action `yaml:"actions"`
}
