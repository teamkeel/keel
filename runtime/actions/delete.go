package actions

import (
	"errors"
	"fmt"

	"github.com/teamkeel/keel/proto"
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

	if action.scope.operation.Implementation == proto.OperationImplementation_OPERATION_IMPLEMENTATION_CUSTOM {
		client := action.scope.customFunctionClient
		res, err := client.Call(action.scope.context, action.scope.operation.Name, action.scope.operation.Type, action.scope.writeValues)

		if err != nil {
			return nil, err
		}

		resMap, ok := res.(map[string]any)

		if !ok {
			return nil, fmt.Errorf("not a map")
		}

		object, ok := resMap["success"]

		if !ok {
			return nil, fmt.Errorf("no success key")
		}

		success, ok := object.(bool)

		if !ok {
			return nil, fmt.Errorf("success not a bool")
		}

		return &ActionResult[DeleteResult]{
			Value: DeleteResult{
				Success: success,
			},
		}, nil
	}

	records := []map[string]any{}
	err := action.scope.query.Delete(records).Error

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
