package executor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMockAccumulatesActions(t *testing.T) {
	m := NewMock()
	ctx := context.Background()

	require.NoError(t, m.EnsureRole(ctx, "role1"))
	require.NoError(t, m.DeleteRole(ctx, "role2"))
	require.NoError(t, m.AddPermissions(ctx, "role1", []string{"a", "b"}))
	require.NoError(t, m.RemovePermissions(ctx, "role1", []string{"c"}))

	require.ElementsMatch(t, []string{"role1"}, m.RolesEnsured)
	require.ElementsMatch(t, []string{"role2"}, m.RolesDeleted)
	require.ElementsMatch(t, []string{"a", "b"}, m.PermsAdded["role1"])
	require.ElementsMatch(t, []string{"c"}, m.PermsRemoved["role1"])
}

func TestMockFailure(t *testing.T) {
	m := NewMock()
	m.FailWith = context.Canceled
	err := m.EnsureRole(context.Background(), "x")
	require.ErrorIs(t, err, context.Canceled)
}
