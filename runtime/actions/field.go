package actions

import (
	"fmt"
)

func Field(fieldName string, source any) (interface{}, error) {
	asMap, ok := source.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("cannot coerce the source object to map[string]any")
	}
	value, ok := asMap[fieldName]
	if !ok {
		return nil, fmt.Errorf("the source map does not contain field: %s", fieldName)
	}
	return value, nil
}
