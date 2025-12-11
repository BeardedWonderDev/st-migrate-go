package schema

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestV1ParserRejectsInvalidEnsure(t *testing.T) {
	input := []byte(`
version: 1
actions:
  - role: admin
    ensure: maybe
`)
	_, err := V1Parser{}.Parse(input)
	require.Error(t, err)
}

func TestV1ParserRejectsEmptyRole(t *testing.T) {
	input := []byte(`
version: 1
actions:
  - role: ""
    ensure: present
`)
	_, err := V1Parser{}.Parse(input)
	require.Error(t, err)
}

func TestRegistryUnknownSchema(t *testing.T) {
	reg := NewRegistry() // empty
	_, err := reg.Parse([]byte(`version: 99`))
	require.Error(t, err)
}
