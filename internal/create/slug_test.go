package create

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSlugify(t *testing.T) {
	require.Equal(t, "add_new_roles", slugify("Add New Roles"))
	require.Equal(t, "a_b_c", slugify("A--B   C"))
}
