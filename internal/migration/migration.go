package migration

// Migration holds both directions for a single version.
type Migration struct {
	Version    uint
	Identifier string
	Up         []byte
	Down       []byte
}
