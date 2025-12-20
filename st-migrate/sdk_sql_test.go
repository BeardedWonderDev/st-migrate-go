package stmigrate

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

var registerErrDriverOnce sync.Once

type errDriver struct{}

func (errDriver) Open(name string) (driver.Conn, error) {
	return nil, errors.New("open failed")
}

func openErrDB(t *testing.T) *sql.DB {
	t.Helper()
	registerErrDriverOnce.Do(func() {
		sql.Register("errdriver-stmigrate", errDriver{})
	})
	db, err := sql.Open("errdriver-stmigrate", "")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestWrapMigrateDatabaseNilDB(t *testing.T) {
	cfg := Config{SourceURL: "file://../testdata/migrations"}
	_, err := NewWithWrappedDatabase(cfg, "postgres", nil, "")
	require.Error(t, err)
}

func TestWrapMigrateDatabaseUnsupportedDriver(t *testing.T) {
	db := &sql.DB{}
	cfg := Config{SourceURL: "file://../testdata/migrations"}
	_, err := NewWithWrappedDatabase(cfg, "oracle", db, "")
	require.Error(t, err)
}

func TestBuildStoreFromDBRequiresDB(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	_, err := buildStoreFromDB("sqlite3", nil, "schema_migrations", false, logger)
	require.Error(t, err)
}

func TestBuildStoreFromSQLiteKeepsDBOpen(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	store, err := buildStoreFromDB("sqlite3", db, "schema_migrations", false, logger)
	require.NoError(t, err)
	require.NotNil(t, store)

	require.NoError(t, store.Close())
	require.NoError(t, db.Ping())
}

func TestBuildStoreFromPostgresConnError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	db := openErrDB(t)

	_, err := buildStoreFromDB("postgres", db, "schema_migrations", false, logger)
	require.Error(t, err)
}

func TestBuildStoreFromMySQLConnError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	db := openErrDB(t)

	_, err := buildStoreFromDB("mysql", db, "schema_migrations", false, logger)
	require.Error(t, err)
}
