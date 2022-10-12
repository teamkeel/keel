package actions

import (
	"errors"
	"fmt"
)

type GetAction struct {
	Action
}

func (action *GetAction) ApplyFilters(args Arguments) ActionBuilder {
	action.query, _ = addGetImplicitInputFilters(action.operation, args, action.query)
	action.query, _ = addGetExplicitInputFilters(action.operation, action.schema, args, action.query)
	return action
}

func (action *GetAction) Execute() (*Result, error) {
	result := []map[string]any{}
	action.query = action.query.Find(&result)

	if action.query.Error != nil {
		return nil, action.query.Error
	}
	n := len(result)
	if n == 0 {
		return nil, errors.New("no records found for Get() operation")
	}
	if n > 1 {
		return nil, fmt.Errorf("Get() operation should find only one record, it found: %d", n)
	}

	var resultMap Result
	resultMap = toLowerCamelMap(result[0])
	return &resultMap, nil
}
