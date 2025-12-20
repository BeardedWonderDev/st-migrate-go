package state

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type stubStore struct {
	versionCalls int
	setCalls     int
	lockCalls    int
	unlockCalls  int
	closeCalls   int
}

func (s *stubStore) Version(_ context.Context) (int, bool, error) {
	s.versionCalls++
	return 1, false, nil
}

func (s *stubStore) SetVersion(_ context.Context, version int, dirty bool) error {
	s.setCalls++
	return nil
}

func (s *stubStore) Lock(_ context.Context) error {
	s.lockCalls++
	return nil
}

func (s *stubStore) Unlock(_ context.Context) error {
	s.unlockCalls++
	return nil
}

func (s *stubStore) Close() error {
	s.closeCalls++
	return nil
}

func TestWrapNoCloseNil(t *testing.T) {
	require.Nil(t, WrapNoClose(nil))
}

func TestNoCloseStoreDelegatesAndSkipsClose(t *testing.T) {
	stub := &stubStore{}
	wrapped := WrapNoClose(stub)
	require.NotNil(t, wrapped)

	ctx := context.Background()
	_, _, err := wrapped.Version(ctx)
	require.NoError(t, err)
	require.NoError(t, wrapped.SetVersion(ctx, 2, false))
	require.NoError(t, wrapped.Lock(ctx))
	require.NoError(t, wrapped.Unlock(ctx))
	require.NoError(t, wrapped.Close())

	require.Equal(t, 1, stub.versionCalls)
	require.Equal(t, 1, stub.setCalls)
	require.Equal(t, 1, stub.lockCalls)
	require.Equal(t, 1, stub.unlockCalls)
	require.Equal(t, 0, stub.closeCalls)
}
