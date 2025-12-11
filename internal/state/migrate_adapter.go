package state

import (
	"context"
	"log/slog"

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
	v, dirty, err := m.driver.Version()
	if err != nil {
		slog.Error("driver version", slog.Any("err", err))
		return 0, false, err
	}
	if v == database.NilVersion {
		v = 0
	}
	return v, dirty, nil
}

func (m *MigrateAdapter) SetVersion(_ context.Context, version int, dirty bool) error {
	if err := m.driver.SetVersion(version, dirty); err != nil {
		slog.Error("driver set version", slog.Int("version", version), slog.Bool("dirty", dirty), slog.Any("err", err))
		return err
	}
	slog.Info("state updated (migrate driver)", slog.Int("version", version), slog.Bool("dirty", dirty))
	return nil
}

func (m *MigrateAdapter) Lock(_ context.Context) error {
	if err := m.driver.Lock(); err != nil {
		slog.Error("driver lock", slog.Any("err", err))
		return err
	}
	slog.Debug("state lock acquired (migrate driver)")
	return nil
}

func (m *MigrateAdapter) Unlock(_ context.Context) error {
	if err := m.driver.Unlock(); err != nil {
		slog.Error("driver unlock", slog.Any("err", err))
		return err
	}
	slog.Debug("state lock released (migrate driver)")
	return nil
}

func (m *MigrateAdapter) Close() error {
	if err := m.driver.Close(); err != nil {
		slog.Error("driver close", slog.Any("err", err))
		return err
	}
	return nil
}
