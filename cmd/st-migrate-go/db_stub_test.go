package main

import (
	"errors"
	"io"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4/database"
	"github.com/stretchr/testify/require"
)

// stubDB implements migrate database.Driver for CLI tests.
type stubDB struct {
	version int
	dirty   bool
	locked  bool
}

func (s *stubDB) Open(url string) (database.Driver, error) { return &stubDB{}, nil }
func (s *stubDB) Close() error                             { return nil }
func (s *stubDB) Lock() error {
	if s.locked {
		return database.ErrLocked
	}
	s.locked = true
	return nil
}
func (s *stubDB) Unlock() error {
	if !s.locked {
		return database.ErrNotLocked
	}
	s.locked = false
	return nil
}
func (s *stubDB) Run(io.Reader) error            { return nil }
func (s *stubDB) SetVersion(v int, d bool) error { s.version, s.dirty = v, d; return nil }
func (s *stubDB) Version() (int, bool, error)    { return s.version, s.dirty, nil }
func (s *stubDB) Drop() error                    { return nil }

func init() {
	// Register stub driver; ignore duplicate panic in case of re-run.
	defer func() { _ = recover() }()
	database.Register("stub", &stubDB{})
}

func TestCLIDatabaseFlagUsesDriver(t *testing.T) {
	// Ensure no file state is created when using database flag.
	tmpDir := t.TempDir()
	stateFile := tmpDir + "/state.json"
	source := "file://" + tmpDir // empty; migrations not needed because no migrations

	cmd := newRootCmd(os.Stdout)
	cmd.SetArgs([]string{"--source", source, "--database", "stub://local", "status"})
	err := cmd.Execute()
	require.NoError(t, err)

	_, err = os.Stat(stateFile)
	require.True(t, errors.Is(err, os.ErrNotExist))
}
