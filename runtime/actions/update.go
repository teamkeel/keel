package actions

import (
	"errors"
	"fmt"
)

type UpdateAction struct {
	scope *Scope
}

type UpdateResult struct {
	Object map[string]any `json:"object"`
}

func (action *UpdateAction) Initialise(scope *Scope) ActionBuilder[UpdateResult] {
	action.scope = scope
	return action
}

// Keep the no-op methods in a group together

func (action *UpdateAction) CaptureImplicitWriteInputValues(args RequestArguments) ActionBuilder[UpdateResult] {
	// Delegate to a method that we hope will become more widely used later.
	if err := DefaultCaptureImplicitWriteInputValues(action.scope.operation.Inputs, args, action.scope); err != nil {
		action.scope.Error = err
		return action
	}
	return action
}

func (action *UpdateAction) CaptureSetValues(args RequestArguments) ActionBuilder[UpdateResult] {
	if err := DefaultCaptureSetValues(action.scope, args); err != nil {
		action.scope.Error = err
		return action
	}
	return action
}

func (action *UpdateAction) IsAuthorised(args RequestArguments) ActionBuilder[UpdateResult] {
	return action // no-op
}

// --------------------

func (action *UpdateAction) ApplyImplicitFilters(args RequestArguments) ActionBuilder[UpdateResult] {
	if action.scope.Error != nil {
		return action
	}
	if err := DefaultApplyImplicitFilters(action.scope, args); err != nil {
		action.scope.Error = err
		return action
	}
	return action
}

func (action *UpdateAction) ApplyExplicitFilters(args RequestArguments) ActionBuilder[UpdateResult] {
	if action.scope.Error != nil {
		return action
	}
	// We delegate to a function that may get used by other Actions later on, once we have
	// unified how we handle operators in both schema where clauses and in implicit inputs language.
	err := DefaultApplyExplicitFilters(action.scope, args)
	if err != nil {
		action.scope.Error = err
		return action
	}
	return action
}

func (action *UpdateAction) Execute(args RequestArguments) (*ActionResult[UpdateResult], error) {
	result := []map[string]any{}
	action.scope.query = action.scope.query.Find(&result)
	if action.scope.query.Error != nil {
		return nil, action.scope.query.Error
	}
	n := len(result)
	if n == 0 {
		return nil, errors.New("no records found for Update() operation")
	}
	if n > 1 {
		return nil, fmt.Errorf("Update() operation should find only one record, it found: %d", n)
	}
	authorized, err := EvaluatePermissions(action.scope.context, action.scope.operation, action.scope.schema, toLowerCamelMap(result[0]))
	if err != nil {
		return nil, err
	}
	if !authorized {
		return nil, errors.New("not authorized to access this operation")
	}

	err = action.scope.query.Updates(action.scope.writeValues).Error

	if err != nil {
		return nil, err
	}

	return &ActionResult[UpdateResult]{
		Value: UpdateResult{
			Object: toLowerCamelMap(action.scope.writeValues),
		},
	}, nil
}
