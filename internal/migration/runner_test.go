package migration

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/BeardedWonderDev/st-migrate-go/internal/executor"
	"github.com/BeardedWonderDev/st-migrate-go/internal/schema"
	"github.com/BeardedWonderDev/st-migrate-go/internal/state/memory"
	"github.com/golang-migrate/migrate/v4/source"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/require"
)

func TestRunnerUpAndDown(t *testing.T) {
	src, err := source.Open("file://../../testdata/migrations")
	require.NoError(t, err)
	migrations, err := LoadAll(src)
	src.Close()
	require.NoError(t, err)
	require.Len(t, migrations, 2)

	store := memory.New()
	exec := executor.NewMock()
	reg := schema.DefaultRegistry()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	r := NewRunner(store, exec, reg, logger, false, migrations)

	ctx := context.Background()
	require.NoError(t, r.Up(ctx, nil))

	version, dirty, err := store.Version(ctx)
	require.NoError(t, err)
	require.False(t, dirty)
	require.Equal(t, 2, version)

	require.Contains(t, exec.RolesEnsured, "app:admin")
	require.Contains(t, exec.RolesEnsured, "app:user")
	require.Contains(t, exec.RolesEnsured, "app:support")
	require.ElementsMatch(t, []string{"app:read", "app:write"}, exec.PermsAdded["app:admin"])

	// Roll back one step
	require.NoError(t, r.Down(ctx, 1))
	version, dirty, err = store.Version(ctx)
	require.NoError(t, err)
	require.False(t, dirty)
	require.Equal(t, 1, version)
}

func TestRunnerDirtyStatePreventsActions(t *testing.T) {
	m := memory.New()
	require.NoError(t, m.SetVersion(context.Background(), 1, true))
	exec := executor.NewMock()
	reg := schema.DefaultRegistry()

	r := NewRunner(m, exec, reg, nil, false, []Migration{})
	err := r.Up(context.Background(), nil)
	require.Error(t, err)
	err = r.Down(context.Background(), 1)
	require.Error(t, err)
}

func TestRunnerApplyErrorMarksDirty(t *testing.T) {
	src, err := source.Open("file://../../testdata/migrations")
	require.NoError(t, err)
	migrations, err := LoadAll(src)
	src.Close()
	require.NoError(t, err)

	store := memory.New()
	exec := executor.NewMock()
	exec.FailWith = context.Canceled
	reg := schema.DefaultRegistry()
	r := NewRunner(store, exec, reg, nil, false, migrations)

	err = r.Up(context.Background(), nil)
	require.Error(t, err)
	_, dirty, derr := store.Version(context.Background())
	require.NoError(t, derr)
	require.True(t, dirty)
}

func TestRunnerMissingDownErrors(t *testing.T) {
	store := memory.New()
	require.NoError(t, store.SetVersion(context.Background(), 5, false))
	exec := executor.NewMock()
	reg := schema.DefaultRegistry()
	r := NewRunner(store, exec, reg, nil, false, []Migration{
		{Version: 1, Up: []byte("version:1"), Down: []byte("version:1")},
	})
	err := r.Down(context.Background(), 1)
	require.Error(t, err)
}

func TestRunnerUpRespectsTarget(t *testing.T) {
	src, err := source.Open("file://../../testdata/migrations")
	require.NoError(t, err)
	migrations, err := LoadAll(src)
	src.Close()
	require.NoError(t, err)

	store := memory.New()
	exec := executor.NewMock()
	reg := schema.DefaultRegistry()
	r := NewRunner(store, exec, reg, nil, false, migrations)

	target := uint(1)
	require.NoError(t, r.Up(context.Background(), &target))

	v, dirty, err := store.Version(context.Background())
	require.NoError(t, err)
	require.False(t, dirty)
	require.Equal(t, 1, v)
}

func TestRunnerHandlesNoMigrations(t *testing.T) {
	store := memory.New()
	exec := executor.NewMock()
	reg := schema.DefaultRegistry()
	r := NewRunner(store, exec, reg, nil, false, []Migration{})
	require.NoError(t, r.Up(context.Background(), nil))
	v, dirty, err := store.Version(context.Background())
	require.NoError(t, err)
	require.Equal(t, 0, v)
	require.False(t, dirty)
}
