package file

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFileStorePersistsAndLocks(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")

	store, err := New(path)
	require.NoError(t, err)

	// initial version defaults to 0, clean
	v, dirty, err := store.Version(context.Background())
	require.NoError(t, err)
	require.Equal(t, 0, v)
	require.False(t, dirty)

	require.NoError(t, store.SetVersion(context.Background(), 3, true))
	v, dirty, err = store.Version(context.Background())
	require.NoError(t, err)
	require.Equal(t, 3, v)
	require.True(t, dirty)

	// lock / unlock cycle
	require.NoError(t, store.Lock(context.Background()))
	require.ErrorIs(t, store.Lock(context.Background()), ErrLocked)
	require.NoError(t, store.Unlock(context.Background()))
	require.ErrorIs(t, store.Unlock(context.Background()), ErrNotLocked)

	// persisted to disk
	require.FileExists(t, path)
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Contains(t, string(raw), `"version": 3`)

	require.NoError(t, store.Close())
}

func TestFileStoreCreateMissingDirs(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "state")
	path := filepath.Join(dir, "state.json")
	store, err := New(path)
	require.NoError(t, err)
	require.NoError(t, store.SetVersion(context.Background(), 1, false))
	require.FileExists(t, path)
}

func TestFileStoreInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
	require.NoError(t, os.WriteFile(path, []byte("{bad json"), 0o644))

	store, err := New(path)
	require.NoError(t, err)
	_, _, err = store.Version(context.Background())
	require.Error(t, err)
}

func TestFileStoreWriteFailsOnDirectoryPath(t *testing.T) {
	dir := t.TempDir()
	store, err := New(dir)
	require.NoError(t, err)

	err = store.SetVersion(context.Background(), 1, false)
	require.Error(t, err)
}
