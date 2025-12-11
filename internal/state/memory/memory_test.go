package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMemoryStoreVersionAndLocks(t *testing.T) {
	store := New()
	v, dirty, err := store.Version(context.Background())
	require.NoError(t, err)
	require.Equal(t, 0, v)
	require.False(t, dirty)

	require.NoError(t, store.SetVersion(context.Background(), 2, true))
	v, dirty, err = store.Version(context.Background())
	require.NoError(t, err)
	require.Equal(t, 2, v)
	require.True(t, dirty)

	require.NoError(t, store.Lock(context.Background()))
	require.ErrorIs(t, store.Lock(context.Background()), ErrLocked)
	require.NoError(t, store.Unlock(context.Background()))
	require.ErrorIs(t, store.Unlock(context.Background()), ErrNotLocked)
}
