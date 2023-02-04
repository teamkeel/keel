package runtimetest

import (
	"testing"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/require"
)

// AssertValueAtPath splits a path like "foo.bar.baz" into dot-delimited segments (the keys).
// Then it tries to drill into the given map (recursively) using those keys.
// It then makes sure the value thus found is the given expected value.
func AssertValueAtPath(t *testing.T, d map[string]any, path string, expected any) {
	require.Equal(t, expected, GetValueAtPath(t, d, path))
}

// getValueAtPath splits a path like "foo.bar.baz" into dot-delimited segments (the keys).
// Then it tries to drill into the given map (recursively) using those keys.
// It returns the value thus found.
func GetValueAtPath(t *testing.T, theMap map[string]any, path string) any {
	v, _ := jsonpath.Get("$."+path, theMap)
	return v
}

// AssertKSUIDIsNow makes sure that the given value can be
// type-coerced to a ksuid.KSUID, and that the time it encodes
// is roughly now() - given a 5 second tolerance.
func AssertKSUIDIsNow(t *testing.T, v any) {
	s, ok := v.(string)
	require.True(t, ok)
	id, err := ksuid.Parse(s)
	require.NoError(t, err)
	timeSinceMade := time.Since(id.Time())
	require.Less(t, timeSinceMade, 5*time.Second)
}

// AssertIsTimeNow makes sure that the given value can be
// type-coerced to a time.Time, and that its value encodes
// now() - given a 5 second tolerance.
func AssertIsTimeNow(t *testing.T, v any) {
	vTime, ok := v.(time.Time)
	if !ok {
		t.Fatalf("cannot cast this value to a time.Time: %v", v)
	}
	timeSinceMade := time.Since(vTime)
	require.Less(t, timeSinceMade, 5*time.Second)
}
