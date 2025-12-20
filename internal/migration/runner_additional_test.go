package migration

import (
	"context"
	"errors"
	"testing"

	"github.com/BeardedWonderDev/st-migrate-go/internal/executor"
	"github.com/BeardedWonderDev/st-migrate-go/internal/schema"
	"github.com/BeardedWonderDev/st-migrate-go/internal/state/memory"
	"github.com/stretchr/testify/require"
)

type errStore struct {
	lockErr    error
	versionErr error
	setErr     error
	version    int
	dirty      bool
}

func (e *errStore) Version(_ context.Context) (int, bool, error) {
	return e.version, e.dirty, e.versionErr
}

func (e *errStore) SetVersion(_ context.Context, version int, dirty bool) error {
	e.version, e.dirty = version, dirty
	return e.setErr
}

func (e *errStore) Lock(_ context.Context) error   { return e.lockErr }
func (e *errStore) Unlock(_ context.Context) error { return nil }
func (e *errStore) Close() error                   { return nil }

func TestRunnerUpFailsOnLockError(t *testing.T) {
	store := &errStore{lockErr: errors.New("locked")}
	r := NewRunner(store, executor.NewMock(), schema.DefaultRegistry(), nil, false, nil)
	require.Error(t, r.Up(context.Background(), nil))
}

func TestRunnerUpFailsOnVersionError(t *testing.T) {
	store := &errStore{versionErr: errors.New("boom")}
	r := NewRunner(store, executor.NewMock(), schema.DefaultRegistry(), nil, false, nil)
	require.Error(t, r.Up(context.Background(), nil))
}

func TestRunnerDownUsesDefaultStep(t *testing.T) {
	ms := []Migration{
		{Version: 1, Up: []byte("version: 1\nactions:\n  - role: r\n"), Down: []byte("version: 1\nactions:\n  - role: r\n    ensure: absent\n")},
	}
	store := memory.New()
	require.NoError(t, store.SetVersion(context.Background(), 1, false))
	r := NewRunner(store, executor.NewMock(), schema.DefaultRegistry(), nil, false, ms)

	require.NoError(t, r.Down(context.Background(), 0))
	v, dirty, err := store.Version(context.Background())
	require.NoError(t, err)
	require.Equal(t, 0, v)
	require.False(t, dirty)
}

func TestRunnerUpNormalizesNegativeVersion(t *testing.T) {
	ms := []Migration{
		{Version: 1, Up: []byte("version: 1\nactions:\n  - role: r\n"), Down: []byte("version: 1\nactions:\n  - role: r\n    ensure: absent\n")},
	}
	store := memory.New()
	require.NoError(t, store.SetVersion(context.Background(), -1, false))
	r := NewRunner(store, executor.NewMock(), schema.DefaultRegistry(), nil, false, ms)

	require.NoError(t, r.Up(context.Background(), nil))
	v, dirty, err := store.Version(context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, v)
	require.False(t, dirty)
}

func TestRunnerUpDryRunDoesNotPersist(t *testing.T) {
	ms := []Migration{
		{Version: 1, Up: []byte("version: 1\nactions:\n  - role: r\n"), Down: []byte("version: 1\nactions:\n  - role: r\n    ensure: absent\n")},
	}
	store := memory.New()
	r := NewRunner(store, executor.NewMock(), schema.DefaultRegistry(), nil, true, ms)

	require.NoError(t, r.Up(context.Background(), nil))
	v, dirty, err := store.Version(context.Background())
	require.NoError(t, err)
	require.Equal(t, 0, v)
	require.False(t, dirty)
}

func TestRunnerDownDryRunDoesNotPersist(t *testing.T) {
	ms := []Migration{
		{Version: 1, Up: []byte("version: 1\nactions:\n  - role: r\n"), Down: []byte("version: 1\nactions:\n  - role: r\n    ensure: absent\n")},
	}
	store := memory.New()
	require.NoError(t, store.SetVersion(context.Background(), 1, false))
	r := NewRunner(store, executor.NewMock(), schema.DefaultRegistry(), nil, true, ms)

	require.NoError(t, r.Down(context.Background(), 1))
	v, dirty, err := store.Version(context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, v)
	require.False(t, dirty)
}

func TestRunnerMigrateNoopWhenAtTarget(t *testing.T) {
	ms := []Migration{
		{Version: 1, Up: []byte("version: 1\nactions:\n  - role: r\n"), Down: []byte("version: 1\nactions:\n  - role: r\n    ensure: absent\n")},
	}
	store := memory.New()
	require.NoError(t, store.SetVersion(context.Background(), 1, false))
	r := NewRunner(store, executor.NewMock(), schema.DefaultRegistry(), nil, false, ms)

	require.NoError(t, r.Migrate(context.Background(), 1))
	v, dirty, err := store.Version(context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, v)
	require.False(t, dirty)
}

func TestRunnerMigrateUpToSetVersionError(t *testing.T) {
	ms := []Migration{
		{Version: 1, Up: []byte("version: 1\nactions:\n  - role: r\n"), Down: []byte("version: 1\nactions:\n  - role: r\n    ensure: absent\n")},
	}
	store := &errStore{setErr: errors.New("set version")}
	r := NewRunner(store, executor.NewMock(), schema.DefaultRegistry(), nil, false, ms)

	require.Error(t, r.Migrate(context.Background(), 1))
}

func TestRunnerMigrateDownDryRunDoesNotPersist(t *testing.T) {
	ms := []Migration{
		{Version: 1, Up: []byte("version: 1\nactions:\n  - role: r\n"), Down: []byte("version: 1\nactions:\n  - role: r\n    ensure: absent\n")},
		{Version: 2, Up: []byte("version: 1\nactions:\n  - role: r2\n"), Down: []byte("version: 1\nactions:\n  - role: r2\n    ensure: absent\n")},
	}
	store := memory.New()
	require.NoError(t, store.SetVersion(context.Background(), 2, false))
	r := NewRunner(store, executor.NewMock(), schema.DefaultRegistry(), nil, true, ms)

	require.NoError(t, r.Migrate(context.Background(), 0))
	v, dirty, err := store.Version(context.Background())
	require.NoError(t, err)
	require.Equal(t, 2, v)
	require.False(t, dirty)
}

func TestRunnerApplyRemovesPermissions(t *testing.T) {
	exec := executor.NewMock()
	r := NewRunner(memory.New(), exec, schema.DefaultRegistry(), nil, false, nil)
	data := []byte("version: 1\nactions:\n  - role: r\n    ensure: present\n    remove:\n      - p1\n")

	require.NoError(t, r.apply(context.Background(), 1, data))
	require.Equal(t, []string{"p1"}, exec.PermsRemoved["r"])
}

func TestRunnerApplyAddPermissionsError(t *testing.T) {
	exec := executor.NewMock()
	exec.FailWith = errors.New("boom")
	r := NewRunner(memory.New(), exec, schema.DefaultRegistry(), nil, false, nil)
	data := []byte("version: 1\nactions:\n  - role: r\n    ensure: present\n    add:\n      - p1\n")

	require.Error(t, r.apply(context.Background(), 1, data))
}
