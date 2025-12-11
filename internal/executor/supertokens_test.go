package executor

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	userrolesmodels "github.com/supertokens/supertokens-golang/recipe/userroles/userrolesmodels"
	supertokens "github.com/supertokens/supertokens-golang/supertokens"
)

type mockRolesClient struct {
	calls []string
	fail  error
	resp  userrolesmodels.DeleteRoleResponse
}

func (m *mockRolesClient) CreateNewRoleOrAddPermissions(role string, perms []string, ctx supertokens.UserContext) (userrolesmodels.CreateNewRoleOrAddPermissionsResponse, error) {
	m.calls = append(m.calls, "add:"+role)
	return userrolesmodels.CreateNewRoleOrAddPermissionsResponse{}, m.fail
}
func (m *mockRolesClient) RemovePermissionsFromRole(role string, perms []string, ctx supertokens.UserContext) (userrolesmodels.RemovePermissionsFromRoleResponse, error) {
	m.calls = append(m.calls, "remove:"+role)
	return userrolesmodels.RemovePermissionsFromRoleResponse{}, m.fail
}
func (m *mockRolesClient) DeleteRole(role string, ctx supertokens.UserContext) (userrolesmodels.DeleteRoleResponse, error) {
	m.calls = append(m.calls, "delete:"+role)
	return m.resp, m.fail
}

func TestSuperTokensExecutorUsesClient(t *testing.T) {
	mock := &mockRolesClient{resp: userrolesmodels.DeleteRoleResponse{OK: &struct{ DidRoleExist bool }{DidRoleExist: true}}}
	OverrideRolesClient(mock)
	defer OverrideRolesClient(nil)

	exec := NewSuperTokensExecutor()
	ctx := context.Background()

	require.NoError(t, exec.EnsureRole(ctx, "role"))
	require.NoError(t, exec.AddPermissions(ctx, "role", []string{"p"}))
	require.NoError(t, exec.RemovePermissions(ctx, "role", []string{"p"}))
	require.NoError(t, exec.DeleteRole(ctx, "role"))

	require.Len(t, mock.calls, 4)
}

func TestSuperTokensExecutorHandlesEmptyInputs(t *testing.T) {
	mock := &mockRolesClient{}
	OverrideRolesClient(mock)
	defer OverrideRolesClient(nil)
	exec := NewSuperTokensExecutor()
	ctx := context.Background()

	require.NoError(t, exec.EnsureRole(ctx, ""))
	require.NoError(t, exec.AddPermissions(ctx, "role", nil))
	require.NoError(t, exec.RemovePermissions(ctx, "role", nil))
}

func TestSuperTokensExecutorPropagatesErrors(t *testing.T) {
	mock := &mockRolesClient{fail: errors.New("boom")}
	OverrideRolesClient(mock)
	defer OverrideRolesClient(nil)
	exec := NewSuperTokensExecutor()
	err := exec.EnsureRole(context.Background(), "r")
	require.Error(t, err)
}
