package state

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/golang-migrate/migrate/v4/database"
	"github.com/stretchr/testify/require"
)

type stubDriver struct {
	version    int
	dirty      bool
	locked     bool
	closeCalls int
}

func (s *stubDriver) Open(url string) (database.Driver, error) { return nil, errors.New("not implemented") }
func (s *stubDriver) Close() error                            { s.closeCalls++; return nil }
func (s *stubDriver) Lock() error {
	if s.locked {
		return database.ErrLocked
	}
	s.locked = true
	return nil
}
func (s *stubDriver) Unlock() error {
	if !s.locked {
		return database.ErrNotLocked
	}
	s.locked = false
	return nil
}
func (s *stubDriver) Run(_ io.Reader) error                    { return nil }
func (s *stubDriver) SetVersion(v int, d bool) error           { s.version, s.dirty = v, d; return nil }
func (s *stubDriver) Version() (int, bool, error)              { return s.version, s.dirty, nil }
func (s *stubDriver) Drop() error                              { return nil }

func TestMigrateAdapterDelegates(t *testing.T) {
	stub := &stubDriver{}
	adapter := NewMigrateAdapter(stub)

	require.NoError(t, adapter.SetVersion(context.Background(), 5, true))
	v, dirty, err := adapter.Version(context.Background())
	require.NoError(t, err)
	require.Equal(t, 5, v)
	require.True(t, dirty)

	require.NoError(t, adapter.Lock(context.Background()))
	require.ErrorIs(t, adapter.Lock(context.Background()), database.ErrLocked)
	require.NoError(t, adapter.Unlock(context.Background()))
	require.ErrorIs(t, adapter.Unlock(context.Background()), database.ErrNotLocked)

	require.NoError(t, adapter.Close())
	require.Equal(t, 1, stub.closeCalls)
}
