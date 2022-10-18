package actions

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

func (action *UpdateAction) ApplyExplicitFilters(args RequestArguments) ActionBuilder[UpdateResult] {
	return action // no-op
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

func (action *UpdateAction) Execute(args RequestArguments) (*ActionResult[UpdateResult], error) {
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
