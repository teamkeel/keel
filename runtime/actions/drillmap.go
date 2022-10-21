package actions

import (
	"errors"
	"fmt"
)

// DrillMap navigates keys in a hierachical map of arbitrary depth. For example if you pass in keys like:
// ["foo", "bar", "baz"] it first looks up "foo" in the given map. It expects the value there also to be
// a map[string]any,
// and proceeds to look up the next key "bar" key in that map. And so on recursively.
// It returns the value thus found. It copes safely with the hierarchy not being composed properly with the
// correctly typed maps, and it copes with a key lookup failing at each level. In other words it will
// not panic.
func DrillMap(m map[string]any, keys []string) (any, error) {
	if len(keys) == 0 {
		return nil, errors.New("invalid argument - empty list of keys")
	}

	// Attempt to look up the next key.
	v, ok := m[keys[0]]
	if !ok {
		return nil, fmt.Errorf("this map: %+v, does not contain the key: %s", m, keys[0])
	}

	// If we've reached (recursed to) the final key - we can just return the corresponding value.
	if len(keys) == 1 {
		return v, nil
	}

	// Otherwise, we require that v is a subMap, and recurse using the sub map,
	// and the remaining keys.
	subMap, ok := v.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("cannot cast this value: %v to a map[string]any", v)
	}
	remainingPath := keys[1:]
	return DrillMap(subMap, remainingPath)
}
