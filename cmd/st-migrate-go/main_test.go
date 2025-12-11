package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/BeardedWonderDev/st-migrate-go/internal/executor"
	stmigrate "github.com/BeardedWonderDev/st-migrate-go/st-migrate"
	"github.com/stretchr/testify/require"
)

func TestCLIUpAndStatusWithFileStore(t *testing.T) {
	stmigrate.SetDefaultExecutorFactory(func() executor.Executor { return executor.NewMock() })

	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "state.json")
	source := "file://" + filepath.Join("..", "..", "testdata", "migrations")

	var out bytes.Buffer
	cmd := newRootCmd(&out)
	cmd.SetArgs([]string{"--source", source, "--state-file", stateFile, "up"})
	require.NoError(t, cmd.Execute())

	out.Reset()
	cmd = newRootCmd(&out)
	cmd.SetArgs([]string{"--source", source, "--state-file", stateFile, "status"})
	require.NoError(t, cmd.Execute())

	require.Contains(t, out.String(), "current version: 2")
	require.Contains(t, out.String(), "pending: none")

	// Ensure state file persisted
	_, err := os.Stat(stateFile)
	require.NoError(t, err)
}

func TestCLICreateWritesFiles(t *testing.T) {
	stmigrate.SetDefaultExecutorFactory(func() executor.Executor { return executor.NewMock() })

	tmpDir := t.TempDir()
	source := "file://" + tmpDir

	var out bytes.Buffer
	cmd := newRootCmd(&out)
	cmd.SetArgs([]string{"--source", source, "--state-file", filepath.Join(tmpDir, "state.json"), "create", "add-logs"})
	require.NoError(t, cmd.Execute())

	entries, err := os.ReadDir(tmpDir)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(entries), 2)
}

func TestCLIDownAndDryRun(t *testing.T) {
	stmigrate.SetDefaultExecutorFactory(func() executor.Executor { return executor.NewMock() })

	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "state.json")
	source := "file://" + filepath.Join("..", "..", "testdata", "migrations")

	var out bytes.Buffer
	cmd := newRootCmd(&out)
	cmd.SetArgs([]string{"--source", source, "--state-file", stateFile, "up"})
	require.NoError(t, cmd.Execute())

	out.Reset()
	cmd = newRootCmd(&out)
	cmd.SetArgs([]string{"--source", source, "--state-file", stateFile, "down", "1"})
	require.NoError(t, cmd.Execute())

	out.Reset()
	cmd = newRootCmd(&out)
	cmd.SetArgs([]string{"--source", source, "--state-file", stateFile, "status"})
	require.NoError(t, cmd.Execute())
	require.Contains(t, out.String(), "current version: 1")

	// dry run should not mutate state
	out.Reset()
	cmd = newRootCmd(&out)
	cmd.SetArgs([]string{"--source", source, "--state-file", stateFile, "--dry-run", "up"})
	require.NoError(t, cmd.Execute())

	// version should remain at 1 after dry-run
	out.Reset()
	cmd = newRootCmd(&out)
	cmd.SetArgs([]string{"--source", source, "--state-file", stateFile, "status"})
	require.NoError(t, cmd.Execute())
	require.Contains(t, out.String(), "current version: 1")
}

func TestCLIUpTargetAppliesPartial(t *testing.T) {
	stmigrate.SetDefaultExecutorFactory(func() executor.Executor { return executor.NewMock() })
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "state.json")
	source := "file://" + filepath.Join("..", "..", "testdata", "migrations")

	var out bytes.Buffer
	cmd := newRootCmd(&out)
	cmd.SetArgs([]string{"--source", source, "--state-file", stateFile, "up", "1"})
	require.NoError(t, cmd.Execute())

	out.Reset()
	cmd = newRootCmd(&out)
	cmd.SetArgs([]string{"--source", source, "--state-file", stateFile, "status"})
	require.NoError(t, cmd.Execute())
	require.Contains(t, out.String(), "current version: 1")
}
