package actions

import (
	"errors"
	"fmt"
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

// --------------------

func (action *DeleteAction) ApplyImplicitFilters(args RequestArguments) ActionBuilder[DeleteResult] {
	if action.scope.Error != nil {
		return action
	}
	if err := DefaultApplyImplicitFilters(action.scope, args); err != nil {
		action.scope.Error = err
		return action
	}
	return action
}

func (action *DeleteAction) ApplyExplicitFilters(args RequestArguments) ActionBuilder[DeleteResult] {
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

func (action *DeleteAction) Execute(args RequestArguments) (*ActionResult[DeleteResult], error) {
	if action.scope.Error != nil {
		return nil, action.scope.Error
	}

	results := []map[string]any{}
	action.scope.query = action.scope.query.Find(&results)
	if action.scope.query.Error != nil {
		return nil, action.scope.query.Error
	}
	n := len(results)
	if n == 0 {
		return nil, errors.New("no records found for Delete() operation")
	}
	if n > 1 {
		return nil, fmt.Errorf("Delete() operation should find only one record, it found: %d", n)
	}

	resultMap := toLowerCamelMap(results[0])

	authorized, err := EvaluatePermissions(action.scope.context, action.scope.operation, action.scope.schema, resultMap)
	if err != nil {
		return nil, err
	}
	if !authorized {
		return nil, errors.New("not authorized to access this operation")
	}

	records := []map[string]any{}
	err = action.scope.query.Delete(records).Error

	result := ActionResult[DeleteResult]{
		Value: DeleteResult{
			Success: err == nil,
		},
	}

	if err != nil {
		return &result, err
	}

	return &result, nil
}

func (action *DeleteAction) IsAuthorised(args RequestArguments) ActionBuilder[DeleteResult] {
	// res := DefaultAuthorizeAction(action.scope, args, action.result)
	return action
}
