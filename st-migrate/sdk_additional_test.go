package stmigrate

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/golang-migrate/migrate/v4/database"
	"github.com/stretchr/testify/require"
)

type stubMigrateDriver struct {
	version int
	dirty   bool
}

func (s *stubMigrateDriver) Open(url string) (database.Driver, error) { return nil, errors.New("not implemented") }
func (s *stubMigrateDriver) Close() error                            { return nil }
func (s *stubMigrateDriver) Lock() error                             { return nil }
func (s *stubMigrateDriver) Unlock() error                           { return nil }
func (s *stubMigrateDriver) Run(_ io.Reader) error                   { return nil }
func (s *stubMigrateDriver) SetVersion(v int, d bool) error          { s.version, s.dirty = v, d; return nil }
func (s *stubMigrateDriver) Version() (int, bool, error)             { return s.version, s.dirty, nil }
func (s *stubMigrateDriver) Drop() error                             { return nil }

func TestNewWithWrappedDriverSuccess(t *testing.T) {
	driver := &stubMigrateDriver{}
	cfg := Config{SourceURL: "file://../testdata/migrations"}

	r, err := NewWithWrappedDriver(cfg, driver)
	require.NoError(t, err)
	t.Cleanup(func() { _ = r.Close() })

	current, pending, err := r.Status(context.Background())
	require.NoError(t, err)
	require.Equal(t, 0, current)
	require.NotEmpty(t, pending)
}

func TestNewErrorsWhenDBDriverWithoutDB(t *testing.T) {
	cfg := Config{SourceURL: "file://../testdata/migrations", DBDriver: "sqlite3"}
	_, err := New(cfg)
	require.Error(t, err)
}
