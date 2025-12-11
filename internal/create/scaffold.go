package create

import (
	"fmt"
	"log/slog"
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
		slog.Error("create migrations directory", slog.String("dir", opts.Dir), slog.Any("err", err))
		return "", "", err
	}
	next, err := nextVersion(opts.Dir)
	if err != nil {
		slog.Error("compute next version", slog.String("dir", opts.Dir), slog.Any("err", err))
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
		slog.Error("write up migration", slog.String("path", upPath), slog.Any("err", err))
		return "", "", err
	}
	if err := os.WriteFile(downPath, []byte(downContent), 0o644); err != nil {
		slog.Error("write down migration", slog.String("path", downPath), slog.Any("err", err))
		return "", "", err
	}
	slog.Info("scaffolded migration pair", slog.String("up", upPath), slog.String("down", downPath), slog.Uint64("version", uint64(next)))
	return upPath, downPath, nil
}

func nextVersion(dir string) (uint, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		slog.Debug("migrations dir missing, starting at 1", slog.String("dir", dir), slog.Any("err", err))
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
