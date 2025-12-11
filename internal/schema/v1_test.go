package schema

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestV1ParserNormalizes(t *testing.T) {
	input := []byte(`
version: 1
actions:
  - role:   " Example:Admin "
    ensure: ""
    add: ["Perm.One", "perm.one", "perm.two "]
    remove: [" x ", "x"]
`)
	parser := V1Parser{}
	spec, err := parser.Parse(input)
	require.NoError(t, err)
	require.Equal(t, 1, spec.Version)
	require.Len(t, spec.Actions, 1)

	act := spec.Actions[0]
	require.Equal(t, "Example:Admin", act.Role)
	require.Equal(t, "present", act.Ensure)
	require.ElementsMatch(t, []string{"perm.one", "perm.two"}, act.Add)
	require.ElementsMatch(t, []string{"x"}, act.Remove)
}
