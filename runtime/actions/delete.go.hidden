package actions

import (
	"fmt"

	"github.com/teamkeel/keel/proto"
)

type DeleteAction struct {
	*Action[DeleteResult]
}

type DeleteResult struct {
	Success bool `json:"success"`
}

func (action *DeleteAction) Initialise(scope *Scope) ActionBuilder[DeleteResult] {
	action.Action = &Action[DeleteResult]{
		Scope: scope,
	}
	return action
}

func (action *DeleteAction) ApplyImplicitFilters(args RequestArguments) ActionBuilder[DeleteResult] {
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

func (action *DeleteAction) Execute(args RequestArguments) (*ActionResult[DeleteResult], error) {
	record := []map[string]any{}
	err := action.query.Delete(record).Error

	result := ActionResult[DeleteResult]{
		Value: DeleteResult{
			Success: err != nil,
		},
	}

	if err != nil {
		return &result, err
	}

	return &result, nil
}
