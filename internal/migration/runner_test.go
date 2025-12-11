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
