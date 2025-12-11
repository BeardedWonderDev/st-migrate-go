package stmigrate

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWrapMigrateDriverNil(t *testing.T) {
	_, err := NewWithWrappedDriver(Config{}, nil)
	require.Error(t, err)
}
