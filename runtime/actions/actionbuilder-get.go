package actions

import (
	"errors"
	"fmt"
)

type GetAction struct {
	Action
}

// func (action *Action) CaptureImplicitWriteInputValues(args RequestArguments) ActionBuilder {
// 	// todo: Default implementation for all actions types
// 	return action
// }

// func (action *Action) CaptureSetValues(args RequestArguments) ActionBuilder {
// 	// todo: Default implementation for all actions types
// 	return action
// }

// func (action *Action) ApplyImplicitFilters(args RequestArguments) ActionBuilder {
// 	// todo: Default implementation for all actions types
// 	return action
// }

// func (action *Action) ApplyExplicitFilters(args RequestArguments) ActionBuilder {
// 	// todo: Default implementation for all actions types
// 	return action
// }

// func (action *Action) IsAuthorised(args RequestArguments) ActionBuilder {
// 	// todo: default implementation for all actions types
// 	return action
// }

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
