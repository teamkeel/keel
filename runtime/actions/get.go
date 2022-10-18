package actions

import (
	"errors"
	"fmt"

	"github.com/teamkeel/keel/proto"
)

type GetAction struct {
	scope *Scope
}

type GetResult struct {
	Object map[string]any `json:"object"`
}

func (action *GetAction) Initialise(scope *Scope) ActionBuilder[GetResult] {
	action.scope = scope
	return action
}

// Keep the no-op methods in a group together

func (action *GetAction) CaptureImplicitWriteInputValues(args RequestArguments) ActionBuilder[GetResult] {
	return action // no-op
}

func (action *GetAction) CaptureSetValues(args RequestArguments) ActionBuilder[GetResult] {
	return action // no-op
}

func (action *GetAction) ApplyExplicitFilters(args RequestArguments) ActionBuilder[GetResult] {
	return action // no-op
}

func (action *GetAction) IsAuthorised(args RequestArguments) ActionBuilder[GetResult] {
	return action // no-op
}

// --------------------

func (action *GetAction) ApplyImplicitFilters(args RequestArguments) ActionBuilder[GetResult] {
	if action.scope.Error != nil {
		return action
	}

	for _, input := range action.scope.operation.Inputs {
		if input.Behaviour != proto.InputBehaviour_INPUT_BEHAVIOUR_IMPLICIT {
			continue
		}

		fieldName := input.Target[0]
		value, ok := args[fieldName]

		if !ok {
			action.scope.Error = fmt.Errorf("this expected input: %s, is missing from this provided args map: %+v", fieldName, args)
			return action
		}

		if err := DRYaddImplicitFilter(action.scope, input, OperatorEquals, value); err != nil {
			action.scope.Error = err
			return action
		}
	}

	return action
}

func (action *GetAction) Execute(args RequestArguments) (*ActionResult[GetResult], error) {
	if action.scope.Error != nil {
		return nil, action.scope.Error
	}

	resultMap := []map[string]any{}
	action.scope.query = action.scope.query.Find(&resultMap)

	if action.scope.query.Error != nil {
		return nil, action.scope.query.Error
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
			Object: resultMap[0],
		},
	}, nil
}
