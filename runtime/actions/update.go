package actions

import (
	"errors"
	"fmt"

	"github.com/teamkeel/keel/proto"
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

		object, ok := resMap["object"]

		if !ok {
			return nil, fmt.Errorf("no object key")
		}

		objectAsMap, ok := object.(map[string]any)

		if !ok {
			return nil, fmt.Errorf("object not a map")
		}

		return &ActionResult[UpdateResult]{
			Value: UpdateResult{
				Object: objectAsMap,
			},
		}, nil
	}

	err := action.scope.query.Updates(action.scope.writeValues).Error

	if err != nil {
		return nil, err
	}

	return &ActionResult[UpdateResult]{
		Value: UpdateResult{
			Object: toLowerCamelMap(action.scope.writeValues),
		},
	}, nil
}
