package executor

import (
	"context"
	"strings"

	"github.com/supertokens/supertokens-golang/recipe/userroles"
	userrolesmodels "github.com/supertokens/supertokens-golang/recipe/userroles/userrolesmodels"
	supertokens "github.com/supertokens/supertokens-golang/supertokens"
)

// RolesClient abstracts the SuperTokens userroles functions for testability.
type RolesClient interface {
	CreateNewRoleOrAddPermissions(role string, perms []string, ctx supertokens.UserContext) (userrolesmodels.CreateNewRoleOrAddPermissionsResponse, error)
	RemovePermissionsFromRole(role string, perms []string, ctx supertokens.UserContext) (userrolesmodels.RemovePermissionsFromRoleResponse, error)
	DeleteRole(role string, ctx supertokens.UserContext) (userrolesmodels.DeleteRoleResponse, error)
}

type superTokensClient struct{}

func (superTokensClient) CreateNewRoleOrAddPermissions(role string, perms []string, ctx supertokens.UserContext) (userrolesmodels.CreateNewRoleOrAddPermissionsResponse, error) {
	return userroles.CreateNewRoleOrAddPermissions(role, perms, ctx)
}
func (superTokensClient) RemovePermissionsFromRole(role string, perms []string, ctx supertokens.UserContext) (userrolesmodels.RemovePermissionsFromRoleResponse, error) {
	return userroles.RemovePermissionsFromRole(role, perms, ctx)
}
func (superTokensClient) DeleteRole(role string, ctx supertokens.UserContext) (userrolesmodels.DeleteRoleResponse, error) {
	return userroles.DeleteRole(role, ctx)
}

var rolesClient RolesClient = superTokensClient{}

// SuperTokensExecutor implements Executor using the SuperTokens roles/permissions API.
type SuperTokensExecutor struct{}

func NewSuperTokensExecutor() *SuperTokensExecutor {
	return &SuperTokensExecutor{}
}

// OverrideRolesClient allows tests to substitute a mock client.
func OverrideRolesClient(client RolesClient) {
	if client == nil {
		rolesClient = superTokensClient{}
	} else {
		rolesClient = client
	}
}

func (s *SuperTokensExecutor) EnsureRole(_ context.Context, role string) error {
	if strings.TrimSpace(role) == "" {
		return nil
	}
	_, err := rolesClient.CreateNewRoleOrAddPermissions(role, []string{}, nil)
	return err
}

func (s *SuperTokensExecutor) DeleteRole(_ context.Context, role string) error {
	if strings.TrimSpace(role) == "" {
		return nil
	}
	resp, err := rolesClient.DeleteRole(role, nil)
	if err != nil {
		return err
	}
	if resp.OK != nil && !resp.OK.DidRoleExist {
		return nil
	}
	return nil
}

func (s *SuperTokensExecutor) AddPermissions(_ context.Context, role string, perms []string) error {
	if len(perms) == 0 {
		return nil
	}
	_, err := rolesClient.CreateNewRoleOrAddPermissions(role, perms, nil)
	return err
}

func (s *SuperTokensExecutor) RemovePermissions(_ context.Context, role string, perms []string) error {
	if len(perms) == 0 {
		return nil
	}
	_, err := rolesClient.RemovePermissionsFromRole(role, perms, nil)
	return err
}
