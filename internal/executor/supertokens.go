package executor

import (
	"context"
	"log/slog"
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
	if err := ensureInitialized(); err != nil {
		return err
	}
	if strings.TrimSpace(role) == "" {
		return nil
	}
	_, err := rolesClient.CreateNewRoleOrAddPermissions(role, []string{}, nil)
	if err != nil {
		slog.Error("supertokens ensure role", slog.String("role", role), slog.Any("err", err))
	} else {
		slog.Info("role ensured", slog.String("role", role))
	}
	return err
}

func (s *SuperTokensExecutor) DeleteRole(_ context.Context, role string) error {
	if err := ensureInitialized(); err != nil {
		return err
	}
	if strings.TrimSpace(role) == "" {
		return nil
	}
	resp, err := rolesClient.DeleteRole(role, nil)
	if err != nil {
		slog.Error("supertokens delete role", slog.String("role", role), slog.Any("err", err))
		return err
	}
	if resp.OK != nil && !resp.OK.DidRoleExist {
		slog.Info("role did not exist; nothing to delete", slog.String("role", role))
		return nil
	}
	slog.Info("role deleted", slog.String("role", role))
	return nil
}

func (s *SuperTokensExecutor) AddPermissions(_ context.Context, role string, perms []string) error {
	if err := ensureInitialized(); err != nil {
		return err
	}
	if len(perms) == 0 {
		return nil
	}
	_, err := rolesClient.CreateNewRoleOrAddPermissions(role, perms, nil)
	if err != nil {
		slog.Error("supertokens add permissions", slog.String("role", role), slog.Any("err", err), slog.Int("count", len(perms)))
	} else {
		slog.Info("permissions added", slog.String("role", role), slog.Int("count", len(perms)))
	}
	return err
}

func (s *SuperTokensExecutor) RemovePermissions(_ context.Context, role string, perms []string) error {
	if err := ensureInitialized(); err != nil {
		return err
	}
	if len(perms) == 0 {
		return nil
	}
	_, err := rolesClient.RemovePermissionsFromRole(role, perms, nil)
	if err != nil {
		slog.Error("supertokens remove permissions", slog.String("role", role), slog.Any("err", err), slog.Int("count", len(perms)))
	} else {
		slog.Info("permissions removed", slog.String("role", role), slog.Int("count", len(perms)))
	}
	return err
}

// ensureInitialized checks that supertokens.Init has been called; if not, returns a helpful error.
func ensureInitialized() error {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("supertokens not initialized; call supertokens.Init before running st-migrate-go")
		}
	}()
	// GetAllCORSHeaders reads the global config and will panic if uninitialized.
	_ = supertokens.GetAllCORSHeaders()
	return nil
}
