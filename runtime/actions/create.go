package actions

import (
	"errors"
)

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

func (action *CreateAction) ApplyExplicitFilters(args RequestArguments) ActionBuilder[CreateResult] {
	return action // no-op
}

func (action *CreateAction) IsCreated(args RequestArguments) ActionBuilder[CreateResult] {
	return action // no-op
}

func (action *CreateAction) ApplyImplicitFilters(args RequestArguments) ActionBuilder[CreateResult] {
	return action // no-op
}

func (action *CreateAction) IsAuthorised(args RequestArguments) ActionBuilder[CreateResult] {
	if action.scope.Error != nil {
		return action
	}

	isAuthorised, err := DefaultIsAuthorised(action.scope, args)

	if err != nil {
		action.scope.Error = err
		return action
	}

	if !isAuthorised {
		action.scope.Error = errors.New("not authorized to access this operation")
	}

	return action
}

func (action *CreateAction) Execute(args RequestArguments) (*ActionResult[CreateResult], error) {
	if action.scope.Error != nil {
		return nil, action.scope.Error
	}

	err := action.scope.query.Create(action.scope.writeValues).Error
	if err != nil {
		action.scope.Error = err
		return nil, err
	}

	result := toLowerCamelMap(action.scope.writeValues)

	return &ActionResult[CreateResult]{
		Value: CreateResult{
			Object: result,
		},
	}, nil
}

func (action *CreateAction) CaptureImplicitWriteInputValues(args RequestArguments) ActionBuilder[CreateResult] {
	if action.scope.Error != nil {
		return action
	}

	// initialise default values
	values, err := initialValueForModel(action.scope.model, action.scope.schema)
	if err != nil {
		action.scope.Error = err
		return action
	}
	action.scope.writeValues = values

	// Delegate to a method that we hope will become more widely used later.
	if err := DefaultCaptureImplicitWriteInputValues(action.scope.operation.Inputs, args, action.scope); err != nil {
		action.scope.Error = err
		return action
	}
	return action
}

func (action *CreateAction) CaptureSetValues(args RequestArguments) ActionBuilder[CreateResult] {
	if action.scope.Error != nil {
		return action
	}

	if err := DefaultCaptureSetValues(action.scope, args); err != nil {
		action.scope.Error = err
		return action
	}
	return action
}
