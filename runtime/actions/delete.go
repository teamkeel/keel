package actions

import (
	"fmt"

	"github.com/teamkeel/keel/proto"
)

type DeleteAction struct {
	scope *Scope
}

type DeleteResult struct {
	Success bool `json:"success"`
}

func (action *DeleteAction) Initialise(scope *Scope) ActionBuilder[DeleteResult] {
	action.scope = scope
	return action
}

// Keep the no-op methods in a group together

func (action *DeleteAction) CaptureImplicitWriteInputValues(args RequestArguments) ActionBuilder[DeleteResult] {
	return action // no-op
}

func (action *DeleteAction) CaptureSetValues(args RequestArguments) ActionBuilder[DeleteResult] {
	return action // no-op
}

func (action *DeleteAction) ApplyExplicitFilters(args RequestArguments) ActionBuilder[DeleteResult] {
	return action // no-op
}

func (action *DeleteAction) IsAuthorised(args RequestArguments) ActionBuilder[DeleteResult] {
	return action // no-op
}

// --------------------

func (action *DeleteAction) ApplyImplicitFilters(args RequestArguments) ActionBuilder[DeleteResult] {
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

func (action *DeleteAction) Execute(args RequestArguments) (*ActionResult[DeleteResult], error) {
	if action.scope.Error != nil {
		return nil, action.scope.Error
	}

	record := []map[string]any{}
	err := action.scope.query.Delete(record).Error

	result := ActionResult[DeleteResult]{
		Value: DeleteResult{
			Success: err != nil,
		},
	}

	if err != nil {
		return nil, err
	}

	return &result, nil
}
