package executor

import (
	"context"
	"sync"
)

// Mock captures applied actions for testing.
type Mock struct {
	mu           sync.Mutex
	RolesEnsured []string
	RolesDeleted []string
	PermsAdded   map[string][]string
	PermsRemoved map[string][]string
	FailWith     error
}

func NewMock() *Mock {
	return &Mock{
		PermsAdded:   map[string][]string{},
		PermsRemoved: map[string][]string{},
	}
}

func (m *Mock) EnsureRole(_ context.Context, role string) error {
	if m.FailWith != nil {
		return m.FailWith
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.RolesEnsured = append(m.RolesEnsured, role)
	return nil
}

func (m *Mock) DeleteRole(_ context.Context, role string) error {
	if m.FailWith != nil {
		return m.FailWith
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.RolesDeleted = append(m.RolesDeleted, role)
	return nil
}

func (m *Mock) AddPermissions(_ context.Context, role string, perms []string) error {
	if m.FailWith != nil {
		return m.FailWith
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.PermsAdded[role] = append(m.PermsAdded[role], perms...)
	return nil
}

func (m *Mock) RemovePermissions(_ context.Context, role string, perms []string) error {
	if m.FailWith != nil {
		return m.FailWith
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.PermsRemoved[role] = append(m.PermsRemoved[role], perms...)
	return nil
}
