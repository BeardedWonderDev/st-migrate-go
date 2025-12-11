package stmigrate

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWrapMigrateDriverNil(t *testing.T) {
	require.Nil(t, WrapMigrateDriver(nil))
}
