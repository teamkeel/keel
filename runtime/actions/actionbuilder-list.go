package actions

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/teamkeel/keel/proto"
)

type ListAction struct {
	Action
}

func (action *ListAction) ApplyImplicitFilters(args RequestArguments) ActionBuilder {
	if action.HasError() {
		return action
	}

	allOptional := lo.EveryBy(action.operation.Inputs, func(input *proto.OperationInput) bool {
		return input.Optional
	})

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

			value, ok := whereInputsAsMap[fieldName].(map[string]any)

			// todo: inspect this logic is correct
			if input.Optional && !ok {
				// do not do any further processing if the input is not a map
				// as it is likely nil
				continue
			} else if !ok {
				// not a map, and not optional
				return action.WithError(fmt.Errorf("cannot cast this: %v to a map[string]any", value))
			}

			for operatorStr, operand := range value {
				operatorName, err := operator(operatorStr)
				if err != nil {
					return action.WithError(err)
				}

				action.addImplicitFilter(input, operatorName, operand)
			}
		}
	}

	return action
}

func (action *ListAction) Execute() (*ActionResult, error) {
	return &ActionResult{}, nil
}
