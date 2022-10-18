package actions

import (
	"errors"

	"golang.org/x/exp/maps"
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

func (action *CreateAction) ApplyImplicitFilters(args RequestArguments) ActionBuilder[CreateResult] {
	return action // no-op
}

func (action *CreateAction) IsAuthorised(args RequestArguments) ActionBuilder[CreateResult] {
	return action // no-op
}

func (c *CreateAction) Execute(args RequestArguments) (*ActionResult[CreateResult], error) {
	// initialise default values
	values, err := initialValueForModel(c.scope.model, c.scope.schema)
	if err != nil {
		return nil, err
	}
	maps.Copy(values, c.scope.writeValues)

	// todo: temporary hack for permissions
	authorized, err := EvaluatePermissions(c.scope.context, c.scope.operation, c.scope.schema, toLowerCamelMap(c.scope.writeValues))
	if err != nil {
		return nil, err
	}
	if !authorized {
		return nil, errors.New("not authorized to access this operation")
	}

	err = c.scope.query.Create(values).Error

	if err != nil {
		return nil, err
	}
	result := toLowerCamelMap(values)

	return &ActionResult[CreateResult]{
		Value: CreateResult{
			Object: result,
		},
	}, nil
}

func (action *CreateAction) CaptureImplicitWriteInputValues(args RequestArguments) ActionBuilder[CreateResult] {
	// Delegate to a method that we hope will become more widely used later.
	if err := DefaultCaptureImplicitWriteInputValues(action.scope.operation.Inputs, args, action.scope); err != nil {
		action.scope.Error = err
		return action
	}
	return action
}

func (action *CreateAction) CaptureSetValues(args RequestArguments) ActionBuilder[CreateResult] {
	if err := DefaultCaptureSetValues(action.scope, args); err != nil {
		action.scope.Error = err
		return action
	}
	return action
}
