package actions

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDrillMapHappyPaths(t *testing.T) {
	theMap := map[string]any{
		"foo": map[string]any{
			"bar": map[string]any{
				"baz": 42,
			},
		},
	}
	// General case.
	v, err := DrillMap(theMap, []string{"foo", "bar", "baz"})
	require.Nil(t, err)
	require.Equal(t, 42, v)

	// Two segments
	v, err = DrillMap(theMap, []string{"foo", "bar"})
	require.Nil(t, err)
	asMap, ok := v.(map[string]any)
	require.True(t, ok)
	require.Equal(t, 42, asMap["baz"])

	// One segment
	v, err = DrillMap(theMap, []string{"foo"})
	require.Nil(t, err)
	asMap, ok = v.(map[string]any)
	require.True(t, ok)
	require.Equal(t, map[string]any{
		"baz": 42,
	}, asMap["bar"])
}

func TestDrillErrorNoSuchKey(t *testing.T) {
	theMap := map[string]any{
		"foo": 42,
	}
	_, err := DrillMap(theMap, []string{"bar"})
	require.EqualError(t, err, "this map: map[foo:42], does not contain the key: bar")
}

func TestDrillErrorEmptyKeys(t *testing.T) {
	theMap := map[string]any{}

	// Empty list of keys should cause error.
	_, err := DrillMap(theMap, []string{})
	require.EqualError(t, err, "invalid argument - empty list of keys")
}

func TestDrillErrorMalformedHierarchy(t *testing.T) {
	// Malformed hierarchy of maps should cause error.
	theMap := map[string]any{
		"foo": 42, // Should be a map - given that we drill with two keys.
	}
	_, err := DrillMap(theMap, []string{"foo", "bar"})
	require.EqualError(t, err, "cannot cast this value: 42 to a map[string]any")
}
