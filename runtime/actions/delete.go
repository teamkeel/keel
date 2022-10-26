package actions

import (
	"errors"

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

func (action *DeleteAction) CaptureImplicitWriteInputValues(args ValueArgs) ActionBuilder[DeleteResult] {
	return action // no-op
}

func (action *DeleteAction) CaptureSetValues(args ValueArgs) ActionBuilder[DeleteResult] {
	return action // no-op
}

// --------------------

func (action *DeleteAction) ApplyImplicitFilters(args WhereArgs) ActionBuilder[DeleteResult] {
	if action.scope.Error != nil {
		return action
	}
	if err := DefaultApplyImplicitFilters(action.scope, args); err != nil {
		action.scope.Error = err
		return action
	}
	return action
}

func (action *DeleteAction) ApplyExplicitFilters(args WhereArgs) ActionBuilder[DeleteResult] {
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

func (action *DeleteAction) Execute(args WhereArgs) (*ActionResult[DeleteResult], error) {
	if action.scope.Error != nil {
		return nil, action.scope.Error
	}

	op := action.scope.operation

	if op.Implementation == proto.OperationImplementation_OPERATION_IMPLEMENTATION_CUSTOM {
		return ParseDeleteResponse(action.scope.context, op, args)
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

func (action *DeleteAction) IsAuthorised(args WhereArgs) ActionBuilder[DeleteResult] {
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
