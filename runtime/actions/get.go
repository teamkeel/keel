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

func (action *GetAction) ApplyExplicitFilters(args RequestArguments) ActionBuilder[GetResult] {
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

func (action *GetAction) Execute(args RequestArguments) (*ActionResult[GetResult], error) {
	if action.scope.Error != nil {
		return nil, action.scope.Error
	}

	resultMap := []map[string]any{}
	action.scope.query = action.scope.query.Find(&resultMap)

	if action.scope.query.Error != nil {
		return nil, action.scope.query.Error
	}
	n := len(resultMap)
	if n == 0 {
		return nil, errors.New("no records found for Get() operation")
	}
	if n > 1 {
		return nil, fmt.Errorf("Get() operation should find only one record, it found: %d", n)
	}

	return &ActionResult[GetResult]{
		Value: GetResult{
			Object: resultMap[0],
		},
	}, nil
}
