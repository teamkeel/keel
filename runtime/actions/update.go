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
	if action.scope.Error != nil {
		return action
	}

	// Delegate to a method that we hope will become more widely used later.
	if err := DefaultCaptureImplicitWriteInputValues(action.scope.operation.Inputs, args, action.scope); err != nil {
		action.scope.Error = err
		return action
	}
	return action
}

func (action *UpdateAction) CaptureSetValues(args RequestArguments) ActionBuilder[UpdateResult] {
	if action.scope.Error != nil {
		return action
	}

	if err := DefaultCaptureSetValues(action.scope, args); err != nil {
		action.scope.Error = err
		return action
	}
	return action
}

func (action *UpdateAction) IsAuthorised(args RequestArguments) ActionBuilder[UpdateResult] {
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
	if action.scope.Error != nil {
		return nil, action.scope.Error
	}

	err := action.scope.query.Updates(action.scope.writeValues).Error
	if err != nil {
		return nil, err
	}

	// todo: Use RETURNING statement on UPDATE
	// https://linear.app/keel/issue/RUN-146/gorm-use-returning-on-insert-and-update-statements
	results := []map[string]any{}
	action.scope.query = action.scope.query.Find(&results)

	if action.scope.query.Error != nil {
		return nil, action.scope.query.Error
	}
	n := len(results)
	if n == 0 {
		return nil, errors.New("no records found for Update() operation")
	}
	if n > 1 {
		return nil, fmt.Errorf("Update() operation should find only one record, it found: %d", n)
	}

	result := toLowerCamelMap(results[0])

	return &ActionResult[UpdateResult]{
		Value: UpdateResult{
			Object: result,
		},
	}, nil
}
