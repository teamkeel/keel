package actions

type CreateAction struct {
	scope *Scope
}

type CreateResult struct {
	Object map[string]any `json:"object"`
}

func (action *CreateAction) Initialise(scope *Scope) ActionBuilder[CreateResult] {
	action.scope = scope
	return action
}

// Keep the no-op methods in a group together

func (action *CreateAction) ApplyExplicitFilters(args RequestArguments) ActionBuilder[CreateResult] {
	return action // no-op
}

func (action *CreateAction) ApplyImplicitFilters(args RequestArguments) ActionBuilder[CreateResult] {
	return action // no-op
}

func (action *CreateAction) IsAuthorised(args RequestArguments) ActionBuilder[CreateResult] {
	return action // no-op
}

func (c *CreateAction) Execute(args RequestArguments) (*ActionResult[CreateResult], error) {
	err := c.scope.query.Create(c.scope.writeValues).Error

	if err != nil {
		return nil, err
	}
	result := toLowerCamelMap(c.scope.writeValues)

	return &ActionResult[CreateResult]{
		Value: CreateResult{
			Object: result,
		},
	}, nil
}

func (action *CreateAction) CaptureImplicitWriteInputValues(args RequestArguments) ActionBuilder[CreateResult] {
	if err := DRYCaptureImplicitWriteInputValues(action.scope.operation.Inputs, args, action.scope); err != nil {
		action.scope.Error = err
		return action
	}
	return action
}

func (action *CreateAction) CaptureSetValues(args RequestArguments) ActionBuilder[CreateResult] {
	if err := DRYCaptureSetValues(action.scope, args); err != nil {
		action.scope.Error = err
		return action
	}
	return action
}
