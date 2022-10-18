package actions

import (
	"errors"
	"fmt"
)

type GetAction struct {
	scope *Scope
}

type GetResult struct {
	Object map[string]any `json:"object"`
}

func (action *GetAction) Initialise(scope *Scope) ActionBuilder[GetResult] {
	action.scope = scope
	return action
}

// Keep the no-op methods in a group together

func (action *GetAction) CaptureImplicitWriteInputValues(args RequestArguments) ActionBuilder[GetResult] {
	return action // no-op
}

func (action *GetAction) CaptureSetValues(args RequestArguments) ActionBuilder[GetResult] {
	return action // no-op
}

func (action *GetAction) IsAuthorised(args RequestArguments) ActionBuilder[GetResult] {
	return action // no-op
}

// --------------------

func (action *GetAction) ApplyImplicitFilters(args RequestArguments) ActionBuilder[GetResult] {
	if action.scope.Error != nil {
		return action
	}
	if err := DefaultApplyImplicitFilters(action.scope, args); err != nil {
		action.scope.Error = err
		return action
	}
	return action
}

func (action *GetAction) ApplyExplicitFilters(args RequestArguments) ActionBuilder[GetResult] {
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

func (action *GetAction) Execute(args RequestArguments) (*ActionResult[GetResult], error) {
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
		return nil, errors.New("no records found for Get() operation")
	}
	if n > 1 {
		return nil, fmt.Errorf("Get() operation should find only one record, it found: %d", n)
	}

	singleResult := toLowerCamelMap(results[0])

	// todo: permissions to evaluate at the database-level where applicable
	// https://linear.app/keel/issue/RUN-129/expressions-to-evaluate-at-database-level-where-applicable
	authorized, err := EvaluatePermissions(action.scope.context, action.scope.operation, action.scope.schema, singleResult)
	if err != nil {
		return nil, err
	}
	if !authorized {
		return nil, errors.New("not authorized to access this operation")
	}

	return &ActionResult[GetResult]{
		Value: GetResult{
			Object: toLowerCamelMap(results[0]),
		},
	}, nil
}
