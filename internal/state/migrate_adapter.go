package state

import (
	"context"

	"github.com/golang-migrate/migrate/v4/database"
)

// MigrateAdapter wraps a golang-migrate database.Driver to satisfy the Store interface.
// It delegates version/lock operations and ignores the Run/Drop methods since this
// package handles execution outside of SQL.
type MigrateAdapter struct {
	driver database.Driver
}

// NewMigrateAdapter constructs a Store backed by a migrate database driver.
func NewMigrateAdapter(driver database.Driver) *MigrateAdapter {
	return &MigrateAdapter{driver: driver}
}

func (m *MigrateAdapter) Version(_ context.Context) (int, bool, error) {
	return m.driver.Version()
}

func (m *MigrateAdapter) SetVersion(_ context.Context, version int, dirty bool) error {
	return m.driver.SetVersion(version, dirty)
}

func (m *MigrateAdapter) Lock(_ context.Context) error {
	return m.driver.Lock()
}

func (m *MigrateAdapter) Unlock(_ context.Context) error {
	return m.driver.Unlock()
}

func (m *MigrateAdapter) Close() error {
	return m.driver.Close()
}
