package executor

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	userrolesmodels "github.com/supertokens/supertokens-golang/recipe/userroles/userrolesmodels"
	supertokens "github.com/supertokens/supertokens-golang/supertokens"
)

func TestSuperTokensClientDelegates(t *testing.T) {
	prevCreate := createRoleOrAddPermissions
	prevRemove := removePermissionsFromRole
	prevDelete := deleteRole
	t.Cleanup(func() {
		createRoleOrAddPermissions = prevCreate
		removePermissionsFromRole = prevRemove
		deleteRole = prevDelete
	})

	var createRole string
	var createPerms []string
	var createCalled bool
	createRoleOrAddPermissions = func(role string, perms []string, _ ...supertokens.UserContext) (userrolesmodels.CreateNewRoleOrAddPermissionsResponse, error) {
		createCalled = true
		createRole = role
		createPerms = append([]string(nil), perms...)
		return userrolesmodels.CreateNewRoleOrAddPermissionsResponse{}, nil
	}

	var removeRole string
	var removePerms []string
	var removeCalled bool
	removePermissionsFromRole = func(role string, perms []string, _ ...supertokens.UserContext) (userrolesmodels.RemovePermissionsFromRoleResponse, error) {
		removeCalled = true
		removeRole = role
		removePerms = append([]string(nil), perms...)
		return userrolesmodels.RemovePermissionsFromRoleResponse{}, nil
	}

	var deleteRoleName string
	var deleteCalled bool
	deleteRole = func(role string, _ ...supertokens.UserContext) (userrolesmodels.DeleteRoleResponse, error) {
		deleteCalled = true
		deleteRoleName = role
		return userrolesmodels.DeleteRoleResponse{}, nil
	}

	client := superTokensClient{}

	_, err := client.CreateNewRoleOrAddPermissions("role", []string{"a", "b"}, nil)
	require.NoError(t, err)
	_, err = client.RemovePermissionsFromRole("role", []string{"c"}, nil)
	require.NoError(t, err)
	_, err = client.DeleteRole("role", nil)
	require.NoError(t, err)

	require.True(t, createCalled)
	require.Equal(t, "role", createRole)
	require.Equal(t, []string{"a", "b"}, createPerms)

	require.True(t, removeCalled)
	require.Equal(t, "role", removeRole)
	require.Equal(t, []string{"c"}, removePerms)

	require.True(t, deleteCalled)
	require.Equal(t, "role", deleteRoleName)
}

func TestSuperTokensClientPropagatesError(t *testing.T) {
	prevCreate := createRoleOrAddPermissions
	t.Cleanup(func() { createRoleOrAddPermissions = prevCreate })

	want := errors.New("boom")
	createRoleOrAddPermissions = func(role string, perms []string, _ ...supertokens.UserContext) (userrolesmodels.CreateNewRoleOrAddPermissionsResponse, error) {
		return userrolesmodels.CreateNewRoleOrAddPermissionsResponse{}, want
	}

	client := superTokensClient{}
	_, err := client.CreateNewRoleOrAddPermissions("role", nil, nil)
	require.ErrorIs(t, err, want)
}
