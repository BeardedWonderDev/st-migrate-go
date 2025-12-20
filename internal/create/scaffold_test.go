package create

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScaffoldCreatesPairedFiles(t *testing.T) {
	dir := t.TempDir()
	up, down, err := Scaffold(Options{Dir: dir, Name: "Add Logs", Width: 3, SchemaVersion: 2})
	require.NoError(t, err)
	require.FileExists(t, up)
	require.FileExists(t, down)

	require.Contains(t, filepath.Base(up), "001_add_logs.up.yaml")
	require.Contains(t, filepath.Base(down), "001_add_logs.down.yaml")

	data, err := os.ReadFile(up)
	require.NoError(t, err)
	require.Contains(t, string(data), "version: 2")
}

func TestScaffoldIncrementsVersion(t *testing.T) {
	dir := t.TempDir()
	_, _, err := Scaffold(Options{Dir: dir, Name: "first"})
	require.NoError(t, err)
	up2, _, err := Scaffold(Options{Dir: dir, Name: "second"})
	require.NoError(t, err)

	require.Contains(t, filepath.Base(up2), "0002_second.up.yaml")
}

func TestScaffoldDefaultsAndSlugify(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)

	tmp := t.TempDir()
	require.NoError(t, os.Chdir(tmp))
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	up, down, err := Scaffold(Options{Name: "  Mixed Name  "})
	require.NoError(t, err)
	require.FileExists(t, up)
	require.FileExists(t, down)
	require.Contains(t, filepath.Base(up), "0001_mixed_name.up.yaml")

	data, err := os.ReadFile(up)
	require.NoError(t, err)
	require.Contains(t, string(data), "version: 1")
}

func TestScaffoldFailsWhenDirIsFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "not-a-dir")
	require.NoError(t, os.WriteFile(path, []byte("x"), 0o644))

	_, _, err := Scaffold(Options{Dir: path, Name: "bad"})
	require.Error(t, err)
}

func TestNextVersionMissingDirStartsAtOne(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "missing")
	next, err := nextVersion(dir)
	require.NoError(t, err)
	require.Equal(t, uint(1), next)
}
