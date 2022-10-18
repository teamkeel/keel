package actions

import "github.com/teamkeel/keel/proto"

type CreateAction struct {
	scope *Scope
}

type CreateResult struct {
	Object map[string]any `json:"object"`
}

func (action *CreateAction) Initialise(scope *Scope) ActionBuilder[CreateResult] {
	action.scope = scope
	return action
}

// Keep the no-op methods in a group together

func (action *CreateAction) ApplyExplicitFilters(args RequestArguments) ActionBuilder[CreateResult] {
	return action // no-op
}

func (action *CreateAction) ApplyImplicitFilters(args RequestArguments) ActionBuilder[CreateResult] {
	return action // no-op
}

func (action *CreateAction) IsAuthorised(args RequestArguments) ActionBuilder[CreateResult] {
	return action // no-op
}

// --------------

func (c *CreateAction) Execute(args RequestArguments) (*ActionResult[CreateResult], error) {
	err := c.scope.query.Create(c.scope.writeValues).Error

	if err != nil {
		return nil, err
	}
	result := toLowerCamelMap(c.scope.writeValues)

	return &ActionResult[CreateResult]{
		Value: CreateResult{
			Object: result,
		},
	}, nil
}

func (action *CreateAction) CaptureImplicitWriteInputValues(args RequestArguments) ActionBuilder[CreateResult] {
	// Delegate to a method that we hope will become more widely used later.
	if err := captureImplicitWriteInputValues(action.scope.operation.Inputs, args, action.scope); err != nil {
		action.scope.Error = err
		return action
	}
	return action
}

func (action *CreateAction) CaptureSetValues(args RequestArguments) ActionBuilder[CreateResult] {
	if err := DRYCaptureSetValues(action.scope, args); err != nil {
		action.scope.Error = err
		return action
	}
	return action
}

// captureImplicitWriteInputValues updates the writeValues field in the
// given scope object with key/values that represent the implicit write-mode
// inputs carried by the given request.
func captureImplicitWriteInputValues(inputs []*proto.OperationInput, args RequestArguments, scope *Scope) error {

	for _, input := range inputs {
		if input.Behaviour != proto.InputBehaviour_INPUT_BEHAVIOUR_IMPLICIT {
			continue
		}

		if input.Mode != proto.InputMode_INPUT_MODE_WRITE {
			continue
		}

		fieldName := input.Target[0]
		value, ok := args[fieldName]

		if !ok {
			continue
		}

		scope.writeValues[fieldName] = value
	}
	return nil
}
