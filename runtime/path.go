package runtime

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// getValueAtPath splits a path like "foo.bar.baz" into dot-delimited segments (the keys).
// Then it tries to drill into the given map (recursively) using those keys.
// It returns the value thus found.
func getValueAtPath(t *testing.T, theMap map[string]any, path string) any {
	require.NotEqual(t, path, "", "path must not be empty string")
	keys := strings.Split(path, ".")

	v, ok := theMap[keys[0]]
	require.True(t, ok, "this map: %+v, does not contain the key: %s", theMap, keys[0])

	// If we've reached the final key - we can just return the corresponding value.
	if len(keys) == 1 {
		return v
	}
	// Otherwise, we require that v is a subMap, and recurse using the sub map,
	// and the remaining keys.
	subMap, ok := v.(map[string]any)
	require.True(t, ok, "cannot cast this value: %v to a map[string]any", v)
	remainingPath := strings.Join(keys[1:], ".")
	return getValueAtPath(t, subMap, remainingPath)
}
