package actions

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
)

type ListAction struct {
	scope *Scope
}

type ListResult struct {
	Collection  []map[string]any `json:"collection"`
	HasNextPage bool             `json:"hasNextPage"`
}

func (action *ListAction) Initialise(scope *Scope) ActionBuilder[ListResult] {
	action.scope = scope
	return action
}

// Keep the no-op methods in a group together

func (action *ListAction) CaptureImplicitWriteInputValues(args RequestArguments) ActionBuilder[ListResult] {
	return action // no-op
}

func (action *ListAction) CaptureSetValues(args RequestArguments) ActionBuilder[ListResult] {
	return action // no-op
}

func (action *ListAction) IsAuthorised(args RequestArguments) ActionBuilder[ListResult] {
	return action // no-op
}

// ----------------

func (action *ListAction) ApplyImplicitFilters(args RequestArguments) ActionBuilder[ListResult] {
	if action.scope.Error != nil {
		return action
	}

	allOptional := lo.EveryBy(action.scope.operation.Inputs, func(input *proto.OperationInput) bool {
		return input.Optional
	})

inputs:
	for _, input := range action.scope.operation.Inputs {
		if input.Behaviour != proto.InputBehaviour_INPUT_BEHAVIOUR_IMPLICIT {
			continue
		}

		fieldName := input.Target[0]

		whereInputs, ok := args["where"]
		if !ok {
			// We have some required inputs but there is no where key
			if !allOptional {
				action.scope.Error = fmt.Errorf("arguments map does not contain a where key: %v", args)
				return action
			}
		} else {
			whereInputsAsMap, ok := whereInputs.(map[string]any)
			if !ok {
				action.scope.Error = fmt.Errorf("cannot cast this: %v to a map[string]any", whereInputs)
				return action
			}

			value, ok := whereInputsAsMap[fieldName]

			if !ok {
				if input.Optional {
					// do not do any further processing if the input is not a map
					// as it is likely nil
					continue inputs
				}

				action.scope.Error = fmt.Errorf("cannot cast this: %v to a map[string]any", value)
				return action
			}

			valueMap, ok := value.(map[string]any)

			if !ok {
				if input.Optional {
					// do not do any further processing if the input is not a map
					// as it is likely nil
					continue inputs
				}

				action.scope.Error = fmt.Errorf("cannot cast this: %v to a map[string]any", value)
				return action
			}

			for operatorStr, operand := range valueMap {
				operatorName, err := operator(operatorStr) // { "rating": { "greaterThanOrEquals": 1 } }
				if err != nil {
					action.scope.Error = err
					return action
				}

				DRYaddImplicitFilter(action.scope, input, operatorName, operand)
			}
		}
	}

	return action
}

func (action *ListAction) ApplyExplicitFilters(args RequestArguments) ActionBuilder[ListResult] {
	err := DRYApplyExplicitFilters(action.scope, args)
	if err != nil {
		action.scope.Error = err
		return action
	}
	return action
}

func (action *ListAction) Execute(args RequestArguments) (*ActionResult[ListResult], error) {
	// how do we access original args?
	// simple: add

	// pagination:
	// 1. add ordering and lead
	// 2. add after before (pagination)
	// 3. add limit

	// post processing:
	// prune out lead column
	// maybe some other pruning?

	// return return value structure: records, hasNextPage, etc

	panic("hdjds")
}
