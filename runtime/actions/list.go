package actions

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
)

type ListAction struct {
	*Action[ListResult]
}

type ListResult struct {
	Collection  []map[string]any `json:"collection"`
	HasNextPage bool             `json:"hasNextPage"`
}

func (action *ListAction) Initialise(scope *Scope) ActionBuilder[ListResult] {
	action.Action = &Action[ListResult]{
		Scope: scope,
	}
	return action
}

func (action *ListAction) ApplyImplicitFilters(args RequestArguments) ActionBuilder[ListResult] {
	if action.HasError() {
		return action
	}

	allOptional := lo.EveryBy(action.operation.Inputs, func(input *proto.OperationInput) bool {
		return input.Optional
	})

inputs:
	for _, input := range action.operation.Inputs {
		if input.Behaviour != proto.InputBehaviour_INPUT_BEHAVIOUR_IMPLICIT {
			continue
		}

		fieldName := input.Target[0]

		whereInputs, ok := args["where"]
		if !ok {
			// We have some required inputs but there is no where key
			if !allOptional {
				return action.WithError(fmt.Errorf("arguments map does not contain a where key: %v", args))
			}
		} else {
			whereInputsAsMap, ok := whereInputs.(map[string]any)
			if !ok {
				return action.WithError(fmt.Errorf("cannot cast this: %v to a map[string]any", whereInputs))
			}

			value, ok := whereInputsAsMap[fieldName]

			if !ok {
				if input.Optional {
					// do not do any further processing if the input is not a map
					// as it is likely nil
					continue inputs
				}

				return action.WithError(fmt.Errorf("cannot cast this: %v to a map[string]any", value))
			}

			valueMap, ok := value.(map[string]any)

			if !ok {
				if input.Optional {
					// do not do any further processing if the input is not a map
					// as it is likely nil
					continue inputs
				}

				return action.WithError(fmt.Errorf("cannot cast this: %v to a map[string]any", value))
			}

			for operatorStr, operand := range valueMap {
				operatorName, err := operator(operatorStr) // { "rating": { "greaterThanOrEquals": 1 } }
				if err != nil {
					return action.WithError(err)
				}

				action.addImplicitFilter(input, operatorName, operand)
			}
		}
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
