package create

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/golang-migrate/migrate/v4/source"
)

// Options controls creation of new migration files.
type Options struct {
	Dir           string
	Name          string
	Width         int // zero -> default 4
	SchemaVersion int // defaults to 1
}

// Scaffold writes paired up/down YAML files with the next sequential version.
func Scaffold(opts Options) (upPath, downPath string, err error) {
	if opts.Dir == "" {
		opts.Dir = "."
	}
	if opts.Width == 0 {
		opts.Width = 4
	}
	if opts.SchemaVersion == 0 {
		opts.SchemaVersion = 1
	}
	if err := os.MkdirAll(opts.Dir, 0o755); err != nil {
		return "", "", err
	}
	next, err := nextVersion(opts.Dir)
	if err != nil {
		return "", "", err
	}

	slug := slugify(opts.Name)
	filename := fmt.Sprintf("%0*d_%s", opts.Width, next, slug)
	upPath = filepath.Join(opts.Dir, filename+".up.yaml")
	downPath = filepath.Join(opts.Dir, filename+".down.yaml")

	upContent := fmt.Sprintf(`version: %d
actions:
  - role: example:role
    ensure: present
    add: []
    remove: []
`, opts.SchemaVersion)

	downContent := fmt.Sprintf(`version: %d
actions:
  - role: example:role
    ensure: absent
`, opts.SchemaVersion)

	if err := os.WriteFile(upPath, []byte(upContent), 0o644); err != nil {
		return "", "", err
	}
	if err := os.WriteFile(downPath, []byte(downContent), 0o644); err != nil {
		return "", "", err
	}
	return upPath, downPath, nil
}

func nextVersion(dir string) (uint, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 1, nil // directory missing -> start at 1
	}
	var max uint
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		m, err := source.DefaultParse(e.Name())
		if err != nil {
			continue
		}
		if m.Version > max {
			max = m.Version
		}
	}
	return max + 1, nil
}

var slugRE = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = slugRE.ReplaceAllString(name, "_")
	return strings.Trim(name, "_")
}
