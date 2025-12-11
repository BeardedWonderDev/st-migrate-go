package schema

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// V1Parser parses schema version 1 documents.
type V1Parser struct{}

func (V1Parser) Parse(data []byte) (*Spec, error) {
	var spec Spec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("decode schema v1: %w", err)
	}
	if spec.Version == 0 {
		spec.Version = 1
	}
	for i := range spec.Actions {
		action := &spec.Actions[i]
		action.Role = strings.TrimSpace(action.Role)
		if action.Role == "" {
			return nil, fmt.Errorf("schema v1: action %d missing role", i)
		}
		action.Ensure = normalizeEnsure(action.Ensure)
		if action.Ensure != "present" && action.Ensure != "absent" {
			return nil, fmt.Errorf("schema v1: action %s has invalid ensure %q", action.Role, action.Ensure)
		}
		action.Add = normalizePermissions(action.Add)
		action.Remove = normalizePermissions(action.Remove)
	}
	return &spec, nil
}

func normalizeEnsure(val string) string {
	val = strings.TrimSpace(strings.ToLower(val))
	if val == "" {
		return "present"
	}
	return val
}

func normalizePermissions(perms []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(perms))
	for _, p := range perms {
		p = strings.TrimSpace(strings.ToLower(p))
		if p == "" {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return out
}
