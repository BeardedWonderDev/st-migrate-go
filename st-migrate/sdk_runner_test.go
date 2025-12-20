package stmigrate

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/BeardedWonderDev/st-migrate-go/internal/executor"
	"github.com/stretchr/testify/require"
)

func TestSDKUpDownStatus(t *testing.T) {
	SetDefaultExecutorFactory(func() executor.Executor { return executor.NewMock() })
	defer SetDefaultExecutorFactory(nil)

	source := "file://" + filepath.Join("..", "testdata", "migrations")
	r, err := New(Config{SourceURL: source})
	require.NoError(t, err)

	ctx := context.Background()
	current, pending, err := r.Status(ctx)
	require.NoError(t, err)
	require.Equal(t, 0, current)
	require.ElementsMatch(t, []uint{1, 2}, pending)

	require.NoError(t, r.Up(ctx, nil))
	current, pending, err = r.Status(ctx)
	require.NoError(t, err)
	require.Equal(t, 2, current)
	require.Empty(t, pending)

	require.NoError(t, r.Down(ctx, 1))
	current, pending, err = r.Status(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, current)
	require.ElementsMatch(t, []uint{2}, pending)

	require.NoError(t, r.Close())
}
