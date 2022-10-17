package actions

import (
	"errors"
	"fmt"

	"github.com/teamkeel/keel/proto"
)

type GetAction struct {
	Action[GetResult]
}

type GetResult struct {
	Object map[string]any `json:"object"`
}

// func (action *Action) CaptureImplicitWriteInputValues(args RequestArguments) ActionBuilder {
// 	// todo: Default implementation for all actions types
// 	return action
// }

// func (action *Action) CaptureSetValues(args RequestArguments) ActionBuilder {
// 	// todo: Default implementation for all actions types
// 	return action
// }

func (action *GetAction) ApplyImplicitFilters(args RequestArguments) ActionBuilder[GetResult] {
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

func (action *GetAction) Execute(args RequestArguments) (*ActionResult[GetResult], error) {
	resultMap := []map[string]any{}
	action.query = action.query.Find(&resultMap)

	if action.query.Error != nil {
		return nil, action.query.Error
	}
	n := len(resultMap)
	if n == 0 {
		return nil, errors.New("no records found for Get() operation")
	}
	if n > 1 {
		return nil, fmt.Errorf("Get() operation should find only one record, it found: %d", n)
	}

	return &ActionResult[GetResult]{
		Value: GetResult{
			map[string]any{
				"object": resultMap[0],
			},
		},
	}, nil
}
