package executor

import (
	"context"
	"strings"

	"github.com/supertokens/supertokens-golang/recipe/userroles"
)

// SuperTokensExecutor implements Executor using the SuperTokens roles/permissions API.
type SuperTokensExecutor struct{}

func NewSuperTokensExecutor() *SuperTokensExecutor {
	return &SuperTokensExecutor{}
}

func (s *SuperTokensExecutor) EnsureRole(_ context.Context, role string) error {
	if strings.TrimSpace(role) == "" {
		return nil
	}
	_, err := userroles.CreateNewRoleOrAddPermissions(role, []string{}, nil)
	return err
}

func (s *SuperTokensExecutor) DeleteRole(_ context.Context, role string) error {
	if strings.TrimSpace(role) == "" {
		return nil
	}
	ctx := map[string]interface{}{}
	resp, err := userroles.DeleteRole(role, &ctx)
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
	_, err := userroles.CreateNewRoleOrAddPermissions(role, perms, nil)
	return err
}

func (s *SuperTokensExecutor) RemovePermissions(_ context.Context, role string, perms []string) error {
	if len(perms) == 0 {
		return nil
	}
	_, err := userroles.RemovePermissionsFromRole(role, perms, nil)
	return err
}
