package stmigrate

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWrapMigrateDatabaseNil(t *testing.T) {
	require.Nil(t, WrapMigrateDatabase(nil))
}
