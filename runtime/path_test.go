package runtime

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValueAtPath(t *testing.T) {
	theMap := map[string]any{
		"foo": map[string]any{
			"bar": map[string]any{
				"baz": 42,
			},
		},
	}
	// General case.
	v := getValueAtPath(t, theMap, "foo.bar.baz")
	require.Equal(t, 42, v)

	// Two segments
	v = getValueAtPath(t, theMap, "foo.bar")
	asMap, ok := v.(map[string]any)
	require.True(t, ok)
	require.Equal(t, 42, asMap["baz"])

	// One segment
	v = getValueAtPath(t, theMap, "foo")
	asMap, ok = v.(map[string]any)
	require.True(t, ok)
	require.Equal(t, map[string]any{
		"baz": 42,
	}, asMap["bar"])
}
