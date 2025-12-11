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
