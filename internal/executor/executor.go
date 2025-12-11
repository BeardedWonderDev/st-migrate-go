package executor

import "context"

// Executor applies role/permission changes against a backend (e.g., SuperTokens).
type Executor interface {
	EnsureRole(ctx context.Context, role string) error
	DeleteRole(ctx context.Context, role string) error
	AddPermissions(ctx context.Context, role string, perms []string) error
	RemovePermissions(ctx context.Context, role string, perms []string) error
}
