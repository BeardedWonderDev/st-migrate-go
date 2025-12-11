package stmigrate

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/BeardedWonderDev/st-migrate-go/internal/executor"
	"github.com/stretchr/testify/require"
)

func TestNewFailsWithBadSource(t *testing.T) {
	_, err := New(Config{SourceURL: "file:///does/not/exist"})
	require.Error(t, err)
}

func TestSDKMigrateDelegates(t *testing.T) {
	tmp := t.TempDir()
	up := filepath.Join(tmp, "0001_test.up.yaml")
	down := filepath.Join(tmp, "0001_test.down.yaml")
	os.WriteFile(up, []byte("version: 1\nactions:\n  - role: r\n"), 0o644)
	os.WriteFile(down, []byte("version: 1\nactions:\n  - role: r\n    ensure: absent\n"), 0o644)

	SetDefaultExecutorFactory(func() executor.Executor { return executor.NewMock() })
	defer SetDefaultExecutorFactory(nil)

	r, err := New(Config{SourceURL: "file://" + tmp})
	require.NoError(t, err)
	require.NoError(t, r.Migrate(context.Background(), 1))
	require.NoError(t, r.Close())
}

func TestNewUsesDefaultExecutorFactory(t *testing.T) {
	tmp := t.TempDir()
	up := filepath.Join(tmp, "0001_test.up.yaml")
	down := filepath.Join(tmp, "0001_test.down.yaml")
	os.WriteFile(up, []byte("version: 1\nactions:\n  - role: r\n"), 0o644)
	os.WriteFile(down, []byte("version: 1\nactions:\n  - role: r\n  - role: r\n    ensure: absent\n"), 0o644)

	calls := 0
	SetDefaultExecutorFactory(func() executor.Executor {
		calls++
		return executor.NewMock()
	})
	defer SetDefaultExecutorFactory(nil)

	r, err := New(Config{SourceURL: "file://" + tmp})
	require.NoError(t, err)
	require.NotNil(t, r)
	require.Equal(t, 1, calls)
	require.NoError(t, r.Close())
}
