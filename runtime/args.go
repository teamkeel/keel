package runtime

import "fmt"

func toArgsMap(input map[string]any, key string) (map[string]any, error) {
	subKey, ok := input[key]

	if !ok {
		return nil, fmt.Errorf("%s missing", key)
	}

	subMap, ok := subKey.(map[string]any)

	if !ok {
		return nil, fmt.Errorf("%s does not match expected format", key)
	}

	return subMap, nil
}
