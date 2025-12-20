package stmigrate

import (
	"database/sql"
	"io"
	"log/slog"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

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
