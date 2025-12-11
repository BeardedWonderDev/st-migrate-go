package migration

import (
	"context"
	"testing"

	"github.com/BeardedWonderDev/st-migrate-go/internal/executor"
	"github.com/BeardedWonderDev/st-migrate-go/internal/schema"
	"github.com/BeardedWonderDev/st-migrate-go/internal/state/memory"
	"github.com/stretchr/testify/require"
)

func TestRunnerStatusReportsPending(t *testing.T) {
	ms := []Migration{
		{Version: 1, Up: []byte("version:1"), Down: []byte("version:1")},
		{Version: 2, Up: []byte("version:1"), Down: []byte("version:1")},
	}
	store := memory.New()
	require.NoError(t, store.SetVersion(context.Background(), 1, false))
	exec := executor.NewMock()
	reg := schema.DefaultRegistry()
	r := NewRunner(store, exec, reg, nil, false, ms)

	current, pending, err := r.Status(context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, current)
	require.Equal(t, []uint{2}, pending)
}

func TestRunnerMigrateDryRunDoesNotMutate(t *testing.T) {
	ms := []Migration{
		{Version: 1, Up: []byte("version: 1\nactions:\n  - role: r\n"), Down: []byte("version: 1\nactions:\n  - role: r\n    ensure: absent\n")},
	}
	store := memory.New()
	exec := executor.NewMock()
	reg := schema.DefaultRegistry()
	r := NewRunner(store, exec, reg, nil, true, ms)

	require.NoError(t, r.Migrate(context.Background(), 1))
	v, dirty, err := store.Version(context.Background())
	require.NoError(t, err)
	require.Equal(t, 0, v)
	require.False(t, dirty)
}
