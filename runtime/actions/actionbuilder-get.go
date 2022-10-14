package actions

import (
	"errors"
	"fmt"

	"github.com/teamkeel/keel/proto"
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

func (action *GetAction) ApplyImplicitFilters(args RequestArguments) ActionBuilder {
	if action.HasError() {
		return action
	}

	for _, input := range action.operation.Inputs {
		if input.Behaviour != proto.InputBehaviour_INPUT_BEHAVIOUR_IMPLICIT {
			continue
		}

		fieldName := input.Target[0]
		value, ok := args[fieldName]

		if !ok {
			return action.WithError(fmt.Errorf("this expected input: %s, is missing from this provided args map: %+v", fieldName, args))
		}

		action.addImplicitFilter(input, OperatorEquals, value)
	}

	return action
}

// func (action *Action) ApplyExplicitFilters(args RequestArguments) ActionBuilder {
// 	// todo: Default implementation for all actions types
// 	return action
// }

// func (action *Action) IsAuthorised(args RequestArguments) ActionBuilder {
// 	// todo: default implementation for all actions types
// 	return action
// }

func (action *GetAction) Execute() (*ActionResult, error) {
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

	var resultMap ActionResult
	resultMap = toLowerCamelMap(result[0])
	return &resultMap, nil
}
