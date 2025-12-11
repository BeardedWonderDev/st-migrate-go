package stmigrate

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWrapMigrateDatabaseNilDB(t *testing.T) {
	store, err := WrapMigrateDatabase("postgres", nil, "")
	require.Error(t, err)
	require.Nil(t, store)
}

func TestWrapMigrateDatabaseUnsupportedDriver(t *testing.T) {
	db := &sql.DB{}
	store, err := WrapMigrateDatabase("oracle", db, "")
	require.Error(t, err)
	require.Nil(t, store)
}
