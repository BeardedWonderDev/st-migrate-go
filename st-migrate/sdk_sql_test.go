package stmigrate

import (
	"database/sql"
	"testing"

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
