package actions

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

func (action *DeleteAction) ApplyExplicitFilters(args RequestArguments) ActionBuilder[DeleteResult] {
	return action // no-op
}

func (action *DeleteAction) IsAuthorised(args RequestArguments) ActionBuilder[DeleteResult] {
	return action // no-op
}

// --------------------

func (action *DeleteAction) ApplyImplicitFilters(args RequestArguments) ActionBuilder[DeleteResult] {
	if action.scope.Error != nil {
		return action
	}
	if err := applyImplicitFiltersForGetOrDelete(action.scope, args); err != nil {
		action.scope.Error = err
		return action
	}
	return action
}

func (action *DeleteAction) Execute(args RequestArguments) (*ActionResult[DeleteResult], error) {
	if action.scope.Error != nil {
		return nil, action.scope.Error
	}

	record := []map[string]any{}
	err := action.scope.query.Delete(record).Error

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
